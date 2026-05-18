# Feishu Webhook Push Notification for Recharge & Redeem

## Overview

Add Feishu (Lark) webhook push notifications for balance credit events. When a user's balance increases via direct payment recharge or redeem code, the system sends a message card to a configured Feishu webhook URL. The admin can configure enable/disable, webhook URL, which events to notify, and which fields to include in the message card.

## Architecture

```
payment_fulfillment.go          redeem_service.go
  ExecuteBalanceFulfillment        Redeem()
  (直冲到账完成)                    (兑换码使用完成)
       │                                  │
       └──────────┬───────────────────────┘
                  │  go feishuNotifyService.Send(event)
                  ▼
         feishu_notify_service.go
                  │
                  │  buffered channel (cap 64)
                  ▼
         feishuNotifyWorker (goroutine)
                  │
                  │  HTTP POST, timeout 5s
                  ▼
         飞书 Webhook URL
```

- `FeishuNotifyService` holds a buffered channel (cap 64) and a background worker goroutine.
- Callers invoke `go s.Send(event)` — fire-and-forget, zero blocking.
- Worker dequeues events, reloads config from settings table on each send (so admin changes take effect without restart), builds a Feishu message card from the enabled field list, and POSTs to the configured webhook URL with a 5-second timeout.
- On channel full: log warning and drop event (graceful degradation).
- On HTTP failure: log warning only. No retry.

## Data Model

### RechargeEvent (Go)

```go
type RechargeEvent struct {
    UserID        int64   `json:"user_id"`
    UserEmail     string  `json:"user_email"`
    UserName      string  `json:"user_name"`
    Amount        float64 `json:"amount"`
    CreditedAmount float64 `json:"credited_amount"`
    Method        string  `json:"method"`        // "直冲" | "兑换"
    MethodDetail  string  `json:"method_detail"` // "支付宝" / "微信" / "Stripe" / "兑换码"
    OrderNo       string  `json:"order_no"`      // out_trade_no or redeem code
    Time          string  `json:"time"`          // RFC3339
}
```

### Feishu Message Card (interactive msg_type)

```json
{
    "msg_type": "interactive",
    "card": {
        "header": {
            "title": {"tag": "plain_text", "content": "充值到账通知"},
            "template": "blue"
        },
        "elements": [
            {"tag": "div", "text": {"tag": "lark_md", "content": "**用户：**xxx"}},
            {"tag": "div", "text": {"tag": "lark_md", "content": "**金额：**¥100.00"}}
        ]
    }
}
```

The header title switches between "充值到账通知" (direct) and "兑换到账通知" (redeem).
Elements are built dynamically from the admin-configured visible field list.

### Settings Keys (stored in `settings` table)

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `feishu_notify_enabled` | bool | false | Master enable switch |
| `feishu_notify_webhook_url` | string | "" | Feishu webhook URL |
| `feishu_notify_recharge_enabled` | bool | true | Push on direct recharge |
| `feishu_notify_redeem_enabled` | bool | true | Push on redeem code use |
| `feishu_notify_fields` | json | all fields | Visible field names array |

### Field Options

| Field Key | Label | Default Visible |
|-----------|-------|-----------------|
| `user_email` | 用户邮箱 | yes |
| `user_name` | 用户名 | yes |
| `user_id` | 用户ID | no |
| `amount` | 支付金额 | yes |
| `credited_amount` | 实际到账 | yes |
| `method` | 充值方式 | yes |
| `method_detail` | 支付渠道 | yes |
| `order_no` | 订单号/兑换码 | yes |
| `time` | 时间 | yes |

## Backend Changes

### New Files

- `backend/internal/model/feishu_notify.go` — `RechargeEvent` struct, `FeishuNotifyField` constants
- `backend/internal/service/feishu_notify_service.go` — `FeishuNotifyService` with channel, worker, card builder, HTTP sender

### Modified Files

- `backend/internal/service/domain_constants.go` — add 5 setting key constants
- `backend/internal/service/payment_fulfillment.go` — after `markCompleted`, call `go feishuNotify.Send(event)` for balance orders
- `backend/internal/service/redeem_service.go` — after successful redeem, call `go feishuNotify.Send(event)` for balance-type codes
- `backend/internal/service/wire.go` — register `FeishuNotifyService` in ProviderSet
- `backend/internal/handler/dto/settings.go` — add feishu config fields to `SystemSettings` DTO
- `backend/internal/handler/admin/setting_handler.go` — read/store new setting keys in `GetSettings`/`UpdateSettings`

### FeishuNotifyService Core Logic

```go
type FeishuNotifyService struct {
    ch              chan RechargeEvent
    httpClient      *http.Client
    settingRepo     SettingRepository

    mu              sync.RWMutex
    enabled         bool
    webhookURL      string
    rechargeEnabled bool
    redeemEnabled   bool
    visibleFields   []string
}

func NewFeishuNotifyService(settingRepo SettingRepository) *FeishuNotifyService {
    s := &FeishuNotifyService{
        ch:          make(chan RechargeEvent, 64),
        httpClient:  &http.Client{Timeout: 5 * time.Second},
        settingRepo: settingRepo,
    }
    s.reloadConfig(context.Background())
    go s.worker()
    return s
}

func (s *FeishuNotifyService) Send(event RechargeEvent) {
    select {
    case s.ch <- event:
    default:
        log.Warn("feishu notify channel full, dropping event")
    }
}

func (s *FeishuNotifyService) worker() {
    for event := range s.ch {
        ctx := context.Background()
        s.reloadConfig(ctx)
        if !s.enabled { continue }
        if event.Method == "直冲" && !s.rechargeEnabled { continue }
        if event.Method == "兑换" && !s.redeemEnabled { continue }
        card := s.buildCard(event)
        s.httpPost(ctx, card)
    }
}
```

### Integration Points

**In payment_fulfillment.go `ExecuteBalanceFulfillment`:**
After `markCompleted` succeeds, build event:

```go
event := model.RechargeEvent{
    UserID:         order.Edges.User.ID,
    UserEmail:      order.Edges.User.Email,
    UserName:       order.Edges.User.Name,
    Amount:         order.Amount,
    CreditedAmount: order.Amount, // balance orders: amount = credited (fees already deducted by provider)
    Method:          "直冲",
    MethodDetail:    providerName,              // "支付宝" / "微信" / "Stripe"
    OrderNo:         order.OutTradeNo,
    Time:            time.Now().Format(time.RFC3339),
}
go s.feishuNotify.Send(event)
```

**In redeem_service.go `Redeem`:**
After balance redeem succeeds, build event:

```go
event := model.RechargeEvent{
    UserID:         user.ID,
    UserEmail:      user.Email,
    UserName:       user.Name,
    Amount:         redeemValue,
    CreditedAmount: redeemValue,
    Method:          "兑换",
    MethodDetail:    "兑换码",
    OrderNo:         code,
    Time:            time.Now().Format(time.RFC3339),
}
go s.feishuNotify.Send(event)
```

## Frontend Changes

### Modified Files

- `frontend/src/views/admin/SettingsView.vue` — add "通知推送" tab
- `frontend/src/types/index.ts` — add feishu fields to `SystemSettings` type

### Tab Layout

A new "通知推送" tab (icon: bell) in the SettingsView tab list, placed after the "邮件" tab.

Contents:
- **Master toggle**: 启用飞书推送 (switch)
- **Webhook URL**: text input + "测试发送" button (sends a test card to verify configuration)
- **Event toggles**: 直充到账推送 (switch), 兑换到账推送 (switch)
- **Field checkboxes**: 9 checkboxes in a grid, controlling which fields appear in the message card

Wireframe:
```
┌─ 通知推送 ─────────────────────────────┐
│                                          │
│  启用飞书推送          [====○]           │
│                                          │
│  Webhook URL                             │
│  ┌──────────────────────────┐ [测试]    │
│  │ https://open.feishu.cn/..│            │
│  └──────────────────────────┘           │
│                                          │
│  推送事件                                │
│  直充到账推送          [====○]           │
│  兑换到账推送          [====○]           │
│                                          │
│  显示字段                                │
│  ☑ 用户邮箱  ☑ 用户名  ☑ 用户ID        │
│  ☑ 支付金额  ☑ 实际到账 ☑ 充值方式     │
│  ☑ 支付渠道  ☑ 订单号   ☑ 时间         │
│                                          │
└──────────────────────────────────────────┘
```

### Test Send Button

Clicking "测试发送" sends a POST request to a new endpoint `/api/v1/admin/settings/test-feishu` which builds a sample card and sends it to the configured webhook URL. Returns success/failure to the UI for feedback.

A new test endpoint: `POST /api/v1/admin/settings/test-feishu` in `setting_handler.go`.

## Error Handling

- Channel full: log warning, drop event. Notifications are best-effort.
- HTTP timeout (5s): log warning, drop. No retry to avoid backlog.
- HTTP non-2xx: log warning with status code.
- Config parse errors: log error, skip send (treat as disabled).
- Test endpoint: return HTTP 200 with `{success: true/false, message: "..."}` for UI feedback.

## Dependencies

- No new Go modules. Standard library `net/http` for the HTTP call.
- No new npm packages. Existing `Toggle` component and TailwindCSS are sufficient.

## Testing

- Unit tests for `FeishuNotifyService.buildCard` (verify correct field filtering and card structure)
- Unit tests for config reload (verify event filtering by enabled flags)
- Integration test: verify `ExecuteBalanceFulfillment` calls `Send` with correct event fields
- Integration test: verify `Redeem` calls `Send` with correct event fields
- Frontend: verify settings tab renders, toggle switches work, test-send button shows result

## Rollback

Setting `feishu_notify_enabled` to `false` disables all pushes instantly (config reloaded before each send). No code rollback needed to silence notifications.
