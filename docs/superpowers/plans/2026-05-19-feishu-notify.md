# Feishu Webhook Push Notification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Feishu webhook push notifications for direct recharge and redeem code balance credit events, with admin-configurable enable/disable, webhook URL, event types, and visible fields.

**Architecture:** A new `FeishuNotifyService` with a buffered channel + background worker goroutine. Callers in `payment_fulfillment.go` and `redeem_service.go` fire-and-forget events via `go s.Send(event)`. Admin settings are stored in the existing `settings` table and reloaded before each send. The frontend adds a "通知推送" tab in the admin SettingsView.

**Tech Stack:** Go stdlib `net/http`, `encoding/json`. Vue 3 Composition API, existing `Toggle` component, TailwindCSS.

---

## File Structure

### New Files
| File | Responsibility |
|------|---------------|
| `backend/internal/model/feishu_notify.go` | `RechargeEvent` struct, `FeishuNotifyFieldConstants` for field key constants |
| `backend/internal/service/feishu_notify_service.go` | `FeishuNotifyService`: channel, config reload, card builder, HTTP sender |

### Modified Files — Backend
| File | Change |
|------|--------|
| `backend/internal/service/domain_constants.go` | Add 5 setting key constants |
| `backend/internal/service/settings_view.go` | Add `FeishuNotifyEnabled`, `FeishuNotifyWebhookURL`, `FeishuNotifyRechargeEnabled`, `FeishuNotifyRedeemEnabled`, `FeishuNotifyFields` to `SystemSettings` |
| `backend/internal/service/setting_service.go` | Parse and persist the 5 new keys in `parseSettings` and `buildSystemSettingsUpdates` |
| `backend/internal/service/payment_fulfillment.go` | After `markCompleted`, send feishu event for balance orders |
| `backend/internal/service/redeem_service.go` | After balance redeem success, send feishu event |
| `backend/internal/service/wire.go` | Register `FeishuNotifyService` in ProviderSet |
| `backend/internal/handler/dto/settings.go` | Add feishu fields to DTO `SystemSettings` |
| `backend/internal/handler/admin/setting_handler.go` | Add `TestFeishu` handler, read/store feishu keys in settings payload |
| `backend/internal/server/routes/admin.go` | Register `POST /admin/settings/test-feishu` route |
| `backend/cmd/server/wire.go` | No change needed (ProviderSet handles it) |

### Modified Files — Frontend
| File | Change |
|------|--------|
| `frontend/src/api/admin/settings.ts` | Add feishu fields to `SystemSettings` and `UpdateSettingsRequest` interfaces, add `testFeishu` API function |
| `frontend/src/views/admin/SettingsView.vue` | Add "通知推送" tab, form fields, test-send button |
| `frontend/src/i18n/locales/zh-CN.json` | Add i18n keys under `admin.settings.notification` |
| `frontend/src/i18n/locales/en-US.json` | Add i18n keys under `admin.settings.notification` |

---

### Task 1: Create RechargeEvent model and field constants

**Files:**
- Create: `backend/internal/model/feishu_notify.go`

- [ ] **Step 1: Write the model file**

```go
package model

// RechargeEvent holds data for a balance credit notification event.
type RechargeEvent struct {
	UserID         int64   `json:"user_id"`
	UserEmail      string  `json:"user_email"`
	UserName       string  `json:"user_name"`
	Amount         float64 `json:"amount"`
	CreditedAmount float64 `json:"credited_amount"`
	Method         string  `json:"method"`        // "直冲" | "兑换"
	MethodDetail   string  `json:"method_detail"` // "支付宝" / "微信" / "Stripe" / "兑换码"
	OrderNo        string  `json:"order_no"`      // out_trade_no or redeem code
	Time           string  `json:"time"`          // RFC3339
}

// FeishuNotify field key constants for the visible-fields configuration.
const (
	FeishuFieldUserEmail      = "user_email"
	FeishuFieldUserName       = "user_name"
	FeishuFieldUserID         = "user_id"
	FeishuFieldAmount         = "amount"
	FeishuFieldCreditedAmount = "credited_amount"
	FeishuFieldMethod         = "method"
	FeishuFieldMethodDetail   = "method_detail"
	FeishuFieldOrderNo        = "order_no"
	FeishuFieldTime           = "time"
)

// DefaultFeishuNotifyFields returns the default set of visible field keys.
func DefaultFeishuNotifyFields() []string {
	return []string{
		FeishuFieldUserEmail,
		FeishuFieldUserName,
		FeishuFieldAmount,
		FeishuFieldCreditedAmount,
		FeishuFieldMethod,
		FeishuFieldMethodDetail,
		FeishuFieldOrderNo,
		FeishuFieldTime,
	}
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/model/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/model/feishu_notify.go
git commit -m "feat: add RechargeEvent model and feishu notify field constants"
```

---

### Task 2: Add setting key constants to domain_constants.go

**Files:**
- Modify: `backend/internal/service/domain_constants.go`

- [ ] **Step 1: Add the 5 setting key constants**

In `domain_constants.go`, after the existing notification settings block (around `SettingKeyAccountQuotaNotifyEmails`), add:

```go
// Feishu webhook notification
SettingKeyFeishuNotifyEnabled         = "feishu_notify_enabled"
SettingKeyFeishuNotifyWebhookURL      = "feishu_notify_webhook_url"
SettingKeyFeishuNotifyRechargeEnabled = "feishu_notify_recharge_enabled"
SettingKeyFeishuNotifyRedeemEnabled   = "feishu_notify_redeem_enabled"
SettingKeyFeishuNotifyFields          = "feishu_notify_fields"
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/domain_constants.go
git commit -m "feat: add feishu notify setting key constants"
```

---

### Task 3: Add feishu notify fields to internal SystemSettings

**Files:**
- Modify: `backend/internal/service/settings_view.go`

- [ ] **Step 1: Add fields to the `SystemSettings` struct**

After the `AccountQuotaNotifyEmails` field, add:

```go
// Feishu webhook notification
FeishuNotifyEnabled         bool
FeishuNotifyWebhookURL      string
FeishuNotifyRechargeEnabled bool
FeishuNotifyRedeemEnabled   bool
FeishuNotifyFields          []string
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/settings_view.go
git commit -m "feat: add feishu notify fields to internal SystemSettings"
```

---

### Task 4: Implement FeishuNotifyService

**Files:**
- Create: `backend/internal/service/feishu_notify_service.go`

- [ ] **Step 1: Write the service**

```go
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/model"
)

// FeishuNotifyService sends balance credit notifications to a Feishu webhook.
type FeishuNotifyService struct {
	ch          chan model.RechargeEvent
	httpClient  *http.Client
	settingRepo SettingRepository

	mu              sync.RWMutex
	enabled         bool
	webhookURL      string
	rechargeEnabled bool
	redeemEnabled   bool
	visibleFields   []string
}

// NewFeishuNotifyService creates the service and starts the background worker.
func NewFeishuNotifyService(settingRepo SettingRepository) *FeishuNotifyService {
	s := &FeishuNotifyService{
		ch:          make(chan model.RechargeEvent, 64),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		settingRepo: settingRepo,
	}
	s.reloadConfig(context.Background())
	go s.worker()
	return s
}

// Send enqueues an event for asynchronous delivery. Non-blocking.
func (s *FeishuNotifyService) Send(event model.RechargeEvent) {
	select {
	case s.ch <- event:
	default:
		slog.Warn("feishu notify channel full, dropping event")
	}
}

func (s *FeishuNotifyService) worker() {
	for event := range s.ch {
		ctx := context.Background()
		s.reloadConfig(ctx)
		if !s.enabled {
			continue
		}
		if event.Method == "直冲" && !s.rechargeEnabled {
			continue
		}
		if event.Method == "兑换" && !s.redeemEnabled {
			continue
		}
		if s.webhookURL == "" {
			continue
		}
		card := s.buildCard(event)
		if err := s.httpPost(ctx, card); err != nil {
			slog.Warn("feishu notify send failed", "error", err)
		}
	}
}

func (s *FeishuNotifyService) reloadConfig(ctx context.Context) {
	keys := []string{
		SettingKeyFeishuNotifyEnabled,
		SettingKeyFeishuNotifyWebhookURL,
		SettingKeyFeishuNotifyRechargeEnabled,
		SettingKeyFeishuNotifyRedeemEnabled,
		SettingKeyFeishuNotifyFields,
	}
	settings, err := s.settingRepo.GetMultiple(ctx, keys)
	if err != nil {
		slog.Warn("feishu notify config reload failed", "error", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = settings[SettingKeyFeishuNotifyEnabled] == "true"
	s.webhookURL = settings[SettingKeyFeishuNotifyWebhookURL]
	s.rechargeEnabled = settings[SettingKeyFeishuNotifyRechargeEnabled] != "false" // default true
	s.redeemEnabled = settings[SettingKeyFeishuNotifyRedeemEnabled] != "false"     // default true

	if raw := settings[SettingKeyFeishuNotifyFields]; raw != "" {
		var fields []string
		if err := json.Unmarshal([]byte(raw), &fields); err == nil {
			s.visibleFields = fields
		}
	}
	if len(s.visibleFields) == 0 {
		s.visibleFields = model.DefaultFeishuNotifyFields()
	}
}

func (s *FeishuNotifyService) buildCard(event model.RechargeEvent) map[string]any {
	headerTitle := "充值到账通知"
	if event.Method == "兑换" {
		headerTitle = "兑换到账通知"
	}

	fieldSet := make(map[string]bool, len(s.visibleFields))
	for _, f := range s.visibleFields {
		fieldSet[f] = true
	}

	var elements []map[string]any

	maybeAdd := func(key, label, value string) {
		if fieldSet[key] {
			elements = append(elements, map[string]any{
				"tag": "div",
				"text": map[string]string{
					"tag":     "lark_md",
					"content": fmt.Sprintf("**%s：**%s", label, value),
				},
			})
		}
	}

	maybeAdd(model.FeishuFieldUserEmail, "用户邮箱", event.UserEmail)
	maybeAdd(model.FeishuFieldUserName, "用户名", event.UserName)
	maybeAdd(model.FeishuFieldUserID, "用户ID", fmt.Sprintf("%d", event.UserID))
	maybeAdd(model.FeishuFieldAmount, "支付金额", fmt.Sprintf("¥%.2f", event.Amount))
	maybeAdd(model.FeishuFieldCreditedAmount, "实际到账", fmt.Sprintf("¥%.2f", event.CreditedAmount))
	maybeAdd(model.FeishuFieldMethod, "充值方式", event.Method)
	maybeAdd(model.FeishuFieldMethodDetail, "支付渠道", event.MethodDetail)
	maybeAdd(model.FeishuFieldOrderNo, "订单号", event.OrderNo)
	maybeAdd(model.FeishuFieldTime, "时间", event.Time)

	return map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]string{
					"tag":     "plain_text",
					"content": headerTitle,
				},
				"template": "blue",
			},
			"elements": elements,
		},
	}
}

func (s *FeishuNotifyService) httpPost(ctx context.Context, card map[string]any) error {
	body, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/feishu_notify_service.go
git commit -m "feat: add FeishuNotifyService with buffered channel and card builder"
```

---

### Task 5: Register FeishuNotifyService in wire.go

**Files:**
- Modify: `backend/internal/service/wire.go`

- [ ] **Step 1: Add NewFeishuNotifyService to ProviderSet**

In `wire.go`, add `NewFeishuNotifyService` to the `ProviderSet` var (after `ProvideBalanceNotifyService` near line 519):

```go
NewFeishuNotifyService,
```

- [ ] **Step 2: Verify Wire generates correctly**

```bash
cd backend && go generate ./cmd/server
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add backend/internal/service/wire.go backend/cmd/server/wire_gen.go
git commit -m "feat: register FeishuNotifyService in Wire ProviderSet"
```

---

### Task 6: Parse and persist feishu settings in SettingService

**Files:**
- Modify: `backend/internal/service/setting_service.go`

- [ ] **Step 1: Add parsing in `parseSettings`**

After the `AccountQuotaNotifyEmails` parsing, add:

```go
// Feishu notify settings
result.FeishuNotifyEnabled = settings[SettingKeyFeishuNotifyEnabled] == "true"
result.FeishuNotifyWebhookURL = settings[SettingKeyFeishuNotifyWebhookURL]
result.FeishuNotifyRechargeEnabled = settings[SettingKeyFeishuNotifyRechargeEnabled] != "false" // default true
result.FeishuNotifyRedeemEnabled = settings[SettingKeyFeishuNotifyRedeemEnabled] != "false"     // default true
if raw := settings[SettingKeyFeishuNotifyFields]; raw != "" {
	var fields []string
	if err := json.Unmarshal([]byte(raw), &fields); err == nil && len(fields) > 0 {
		result.FeishuNotifyFields = fields
	}
}
if len(result.FeishuNotifyFields) == 0 {
	result.FeishuNotifyFields = model.DefaultFeishuNotifyFields()
}
```

Note: add `"github.com/Wei-Shaw/sub2api/internal/model"` to imports if not already present.

- [ ] **Step 2: Add persistence in `buildSystemSettingsUpdates`**

In the `buildSystemSettingsUpdates` function, after the account quota notification block, add:

```go
// Feishu webhook notification
updates[SettingKeyFeishuNotifyEnabled] = strconv.FormatBool(settings.FeishuNotifyEnabled)
updates[SettingKeyFeishuNotifyWebhookURL] = settings.FeishuNotifyWebhookURL
updates[SettingKeyFeishuNotifyRechargeEnabled] = strconv.FormatBool(settings.FeishuNotifyRechargeEnabled)
updates[SettingKeyFeishuNotifyRedeemEnabled] = strconv.FormatBool(settings.FeishuNotifyRedeemEnabled)
feishuFieldsJSON, err := json.Marshal(settings.FeishuNotifyFields)
if err != nil {
	return nil, fmt.Errorf("marshal feishu notify fields: %w", err)
}
updates[SettingKeyFeishuNotifyFields] = string(feishuFieldsJSON)
```

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/service/setting_service.go
git commit -m "feat: parse and persist feishu notify settings in SettingService"
```

---

### Task 7: Add feishu fields to DTO SystemSettings and UpdateSettingsRequest

**Files:**
- Modify: `backend/internal/handler/dto/settings.go`

- [ ] **Step 1: Add fields to DTO `SystemSettings`**

After `AccountQuotaNotifyEmails`, add:

```go
// Feishu webhook notification
FeishuNotifyEnabled         bool     `json:"feishu_notify_enabled"`
FeishuNotifyWebhookURL      string   `json:"feishu_notify_webhook_url"`
FeishuNotifyRechargeEnabled bool     `json:"feishu_notify_recharge_enabled"`
FeishuNotifyRedeemEnabled   bool     `json:"feishu_notify_redeem_enabled"`
FeishuNotifyFields          []string `json:"feishu_notify_fields"`
```

- [ ] **Step 2: Add fields to `UpdateSettingsRequest` struct in setting_handler.go**

In `backend/internal/handler/admin/setting_handler.go`, after `AffiliateLinkForceBind`, add:

```go
// Feishu webhook notification
FeishuNotifyEnabled         *bool     `json:"feishu_notify_enabled"`
FeishuNotifyWebhookURL      *string   `json:"feishu_notify_webhook_url"`
FeishuNotifyRechargeEnabled *bool     `json:"feishu_notify_recharge_enabled"`
FeishuNotifyRedeemEnabled   *bool     `json:"feishu_notify_redeem_enabled"`
FeishuNotifyFields          *[]string `json:"feishu_notify_fields"`
```

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && go build ./internal/handler/...
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/dto/settings.go backend/internal/handler/admin/setting_handler.go
git commit -m "feat: add feishu notify fields to settings DTO"
```

---

### Task 8: Wire feishu settings in GetSettings and UpdateSettings

**Files:**
- Modify: `backend/internal/handler/admin/setting_handler.go`

- [ ] **Step 1: Add fields to GetSettings response payload**

In `GetSettings`, after the `AffiliateLinkForceBind` line in the `dto.SystemSettings{}` literal, add:

```go
FeishuNotifyEnabled:         settings.FeishuNotifyEnabled,
FeishuNotifyWebhookURL:      settings.FeishuNotifyWebhookURL,
FeishuNotifyRechargeEnabled: settings.FeishuNotifyRechargeEnabled,
FeishuNotifyRedeemEnabled:   settings.FeishuNotifyRedeemEnabled,
FeishuNotifyFields:          settings.FeishuNotifyFields,
```

- [ ] **Step 2: Add merge logic in UpdateSettings**

In the `UpdateSettings` handler, after the `AffiliateLinkForceBind` handling block, add:

```go
// Feishu webhook notification
feishuNotifyEnabled := previousSettings.FeishuNotifyEnabled
if req.FeishuNotifyEnabled != nil {
	feishuNotifyEnabled = *req.FeishuNotifyEnabled
}
feishuNotifyWebhookURL := previousSettings.FeishuNotifyWebhookURL
if req.FeishuNotifyWebhookURL != nil {
	feishuNotifyWebhookURL = *req.FeishuNotifyWebhookURL
}
feishuNotifyRechargeEnabled := previousSettings.FeishuNotifyRechargeEnabled
if req.FeishuNotifyRechargeEnabled != nil {
	feishuNotifyRechargeEnabled = *req.FeishuNotifyRechargeEnabled
}
feishuNotifyRedeemEnabled := previousSettings.FeishuNotifyRedeemEnabled
if req.FeishuNotifyRedeemEnabled != nil {
	feishuNotifyRedeemEnabled = *req.FeishuNotifyRedeemEnabled
}
feishuNotifyFields := previousSettings.FeishuNotifyFields
if req.FeishuNotifyFields != nil {
	feishuNotifyFields = *req.FeishuNotifyFields
}
```

Then pass these values into the `service.SystemSettings` struct when calling `UpdateSettingsWithAuthSourceDefaults`:

```go
FeishuNotifyEnabled:         feishuNotifyEnabled,
FeishuNotifyWebhookURL:      feishuNotifyWebhookURL,
FeishuNotifyRechargeEnabled: feishuNotifyRechargeEnabled,
FeishuNotifyRedeemEnabled:   feishuNotifyRedeemEnabled,
FeishuNotifyFields:          feishuNotifyFields,
```

Note: Find the exact location where `service.SystemSettings{` is constructed inside the `UpdateSettings` handler and add the fields there.

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && go build ./internal/handler/...
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add backend/internal/handler/admin/setting_handler.go
git commit -m "feat: wire feishu notify settings in GetSettings and UpdateSettings"
```

---

### Task 9: Add TestFeishu endpoint

**Files:**
- Modify: `backend/internal/handler/admin/setting_handler.go`
- Modify: `backend/internal/server/routes/admin.go`

- [ ] **Step 1: Add `TestFeishu` handler to SettingHandler**

The SettingHandler needs a reference to `FeishuNotifyService`. Add it to the struct:

```go
type SettingHandler struct {
	settingService        *service.SettingService
	emailService          *service.EmailService
	turnstileService      *service.TurnstileService
	opsService            *service.OpsService
	paymentConfigService  *service.PaymentConfigService
	paymentService        *service.PaymentService
	feishuNotifyService   *service.FeishuNotifyService
}
```

Update `NewSettingHandler` to accept the new dependency:

```go
func NewSettingHandler(
	settingService *service.SettingService,
	emailService *service.EmailService,
	turnstileService *service.TurnstileService,
	opsService *service.OpsService,
	paymentConfigService *service.PaymentConfigService,
	paymentService *service.PaymentService,
	feishuNotifyService *service.FeishuNotifyService,
) *SettingHandler {
```

Add the handler method:

```go
// TestFeishu 发送测试消息到飞书 webhook
// POST /api/v1/admin/settings/test-feishu
func (h *SettingHandler) TestFeishu(c *gin.Context) {
	if h.feishuNotifyService == nil {
		response.BadRequest(c, "feishu notify service not available")
		return
	}
	err := h.feishuNotifyService.SendTestCard(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "test message sent"})
}
```

- [ ] **Step 2: Add `SendTestCard` method to FeishuNotifyService**

In `feishu_notify_service.go`, add:

```go
// SendTestCard sends a sample notification card directly (bypasses channel/worker).
func (s *FeishuNotifyService) SendTestCard(ctx context.Context) error {
	s.reloadConfig(ctx)
	if s.webhookURL == "" {
		return fmt.Errorf("webhook URL is not configured")
	}
	card := s.buildCard(model.RechargeEvent{
		UserEmail:      "test@example.com",
		UserName:       "测试用户",
		UserID:         0,
		Amount:         100.00,
		CreditedAmount: 100.00,
		Method:         "直冲",
		MethodDetail:   "测试",
		OrderNo:        "TEST-001",
		Time:           time.Now().Format(time.RFC3339),
	})
	return s.httpPost(ctx, card)
}
```

- [ ] **Step 3: Register the route**

In `backend/internal/server/routes/admin.go`, in `registerSettingsRoutes`, add after the `send-test-email` route:

```go
adminSettings.POST("/test-feishu", h.Admin.Setting.TestFeishu)
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./...
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/handler/admin/setting_handler.go backend/internal/service/feishu_notify_service.go backend/internal/server/routes/admin.go
git commit -m "feat: add TestFeishu endpoint for testing webhook configuration"
```

---

### Task 10: Integrate into payment fulfillment

**Files:**
- Modify: `backend/internal/service/payment_fulfillment.go`

- [ ] **Step 1: Add feishuNotify field to PaymentService struct**

In `backend/internal/service/payment_service.go` line 172 (`type PaymentService struct`), add after `affiliateService`:

```go
feishuNotify *FeishuNotifyService
```

Add setter method:

```go
func (s *PaymentService) SetFeishuNotify(svc *FeishuNotifyService) {
	s.feishuNotify = svc
}
```

- [ ] **Step 2: Add notification call in doBalance**

In `backend/internal/service/payment_fulfillment.go`, in the `doBalance` method, change line 291 from:

```go
return s.markCompleted(ctx, o, "RECHARGE_SUCCESS")
```

to:

```go
if err := s.markCompleted(ctx, o, "RECHARGE_SUCCESS"); err != nil {
    return err
}
s.sendFeishuRechargeNotify(ctx, o)
return nil
```

And in `redeemActionSkipCompleted` case (line 276), change from:

```go
return s.markCompleted(ctx, o, "RECHARGE_SUCCESS")
```

to:

```go
if err := s.markCompleted(ctx, o, "RECHARGE_SUCCESS"); err != nil {
    return err
}
s.sendFeishuRechargeNotify(ctx, o)
return nil
```

- [ ] **Step 3: Add helper methods in payment_fulfillment.go**

Add at the end of the file:

```go
func (s *PaymentService) sendFeishuRechargeNotify(ctx context.Context, o *dbent.PaymentOrder) {
	if s.feishuNotify == nil {
		return
	}
	user, err := s.entClient.User.Get(ctx, o.UserID)
	if err != nil {
		return
	}
	event := model.RechargeEvent{
		UserID:         user.ID,
		UserEmail:      user.Email,
		UserName:       user.Name,
		Amount:         o.Amount,
		CreditedAmount: o.Amount,
		Method:         "直冲",
		MethodDetail:   s.resolveProviderDisplayName(o),
		OrderNo:        o.OutTradeNo,
		Time:           time.Now().Format(time.RFC3339),
	}
	go s.feishuNotify.Send(event)
}

func (s *PaymentService) resolveProviderDisplayName(o *dbent.PaymentOrder) string {
	switch strings.ToLower(o.PaymentType) {
	case "alipay":
		return "支付宝"
	case "wxpay", "wechat":
		return "微信支付"
	case "stripe":
		return "Stripe"
	case "easypay":
		return "易支付"
	default:
		return o.PaymentType
	}
}
```

Add `"github.com/Wei-Shaw/sub2api/internal/model"` to imports in `payment_fulfillment.go`.

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/payment_fulfillment.go backend/internal/service/payment_service.go
git commit -m "feat: send feishu notification on balance recharge completion"
```

---

### Task 11: Integrate into redeem service

**Files:**
- Modify: `backend/internal/service/redeem_service.go`

- [ ] **Step 1: Add FeishuNotifyService to RedeemService struct**

Add field to `RedeemService`:

```go
type RedeemService struct {
	// ... existing fields ...
	feishuNotify *FeishuNotifyService
}
```

Update `NewRedeemService` to accept it:

```go
func NewRedeemService(
	redeemRepo RedeemCodeRepository,
	userRepo UserRepository,
	subscriptionService *SubscriptionService,
	cache RedeemCache,
	billingCacheService *BillingCacheService,
	entClient *dbent.Client,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	feishuNotify *FeishuNotifyService,
) *RedeemService {
```

- [ ] **Step 2: Add notification call after balance redeem**

In the `Redeem` method, after the `RedeemTypeBalance` case succeeds (after `s.userRepo.UpdateBalance` and before the next case, around line 325), and after transaction commit:

After `s.invalidateRedeemCaches` (around line 370), add:

```go
// Send feishu notification for balance redeem (async, best-effort)
if redeemCode.Type == RedeemTypeBalance && s.feishuNotify != nil {
	event := model.RechargeEvent{
		UserID:         user.ID,
		UserEmail:      user.Email,
		UserName:       user.Name,
		Amount:         redeemCode.Value,
		CreditedAmount: redeemCode.Value,
		Method:         "兑换",
		MethodDetail:   "兑换码",
		OrderNo:        redeemCode.Code,
		Time:           time.Now().Format(time.RFC3339),
	}
	go s.feishuNotify.Send(event)
}
```

Note: Add `"github.com/Wei-Shaw/sub2api/internal/model"` to imports.

- [ ] **Step 3: Check that RedeemService code is after tx.Commit**

The notification must be sent after `tx.Commit()` succeeds, not inside the transaction. Look at the Redeem method flow:
```
tx.Commit()  (line 365)
s.invalidateRedeemCaches(...)  (line 370)
// Add notification here
```

- [ ] **Step 4: Verify it compiles**

```bash
cd backend && go build ./internal/service/
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/redeem_service.go
git commit -m "feat: send feishu notification on redeem code balance credit"
```

---

### Task 12: Wire FeishuNotifyService into PaymentService

**Files:**
- Modify: `backend/internal/service/wire.go`

- [ ] **Step 1: Add ProvidePaymentService wrapper**

`NewPaymentService` is used directly in ProviderSet. Since we added `SetFeishuNotify` as a setter (not a constructor arg), Wire won't inject it automatically. Add a provider function in `wire.go` (e.g., near `ProvideBalanceNotifyService`):

```go
func ProvidePaymentService(
	entClient *dbent.Client,
	registry *payment.Registry,
	loadBalancer payment.LoadBalancer,
	redeemService *RedeemService,
	subscriptionSvc *SubscriptionService,
	configService *PaymentConfigService,
	userRepo UserRepository,
	groupRepo GroupRepository,
	affiliateService *AffiliateService,
	feishuNotify *FeishuNotifyService,
) *PaymentService {
	svc := NewPaymentService(entClient, registry, loadBalancer, redeemService, subscriptionSvc, configService, userRepo, groupRepo, affiliateService)
	svc.SetFeishuNotify(feishuNotify)
	return svc
}
```

Then in ProviderSet, replace `NewPaymentService` with `ProvidePaymentService`.

- [ ] **Step 2: Update RedeemService constructor**

`NewRedeemService` is already in ProviderSet. Since Wire auto-injects based on types, and `FeishuNotifyService` is already in ProviderSet (from Task 5), Wire will pass it to `NewRedeemService` automatically as long as the constructor has the parameter. Verify your updated `NewRedeemService` signature includes `feishuNotify *FeishuNotifyService` as the last parameter.

- [ ] **Step 3: Regenerate Wire**

```bash
cd backend && go generate ./cmd/server
```

Expected: no errors

- [ ] **Step 4: Verify full build**

```bash
cd backend && go build ./...
```

Expected: no errors

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/wire.go backend/cmd/server/wire_gen.go
git commit -m "feat: wire FeishuNotifyService into PaymentService and RedeemService"
```

---

### Task 13: Add feishu fields to frontend SystemSettings type

**Files:**
- Modify: `frontend/src/api/admin/settings.ts`

- [ ] **Step 1: Add fields to `SystemSettings` interface**

After `affiliate_link_force_bind`, add:

```typescript
// Feishu webhook notification
feishu_notify_enabled: boolean;
feishu_notify_webhook_url: string;
feishu_notify_recharge_enabled: boolean;
feishu_notify_redeem_enabled: boolean;
feishu_notify_fields: string[];
```

- [ ] **Step 2: Add fields to `UpdateSettingsRequest` interface**

After `affiliate_link_force_bind?`, add:

```typescript
// Feishu webhook notification
feishu_notify_enabled?: boolean;
feishu_notify_webhook_url?: string;
feishu_notify_recharge_enabled?: boolean;
feishu_notify_redeem_enabled?: boolean;
feishu_notify_fields?: string[];
```

- [ ] **Step 3: Add `testFeishu` API function**

After the `sendTestEmail` function, add:

```typescript
export async function testFeishu(): Promise<{ message: string }> {
  const { data } = await apiClient.post<{ message: string }>(
    "/admin/settings/test-feishu",
  );
  return data;
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/admin/settings.ts
git commit -m "feat: add feishu notify types and testFeishu API to frontend"
```

---

### Task 14: Add i18n keys

**Files:**
- Modify: `frontend/src/i18n/locales/zh-CN.json`
- Modify: `frontend/src/i18n/locales/en-US.json`

- [ ] **Step 1: Add Chinese i18n keys**

Find the `admin.settings` section and add:

```json
"notification": {
  "title": "通知推送",
  "enabled": "启用飞书推送",
  "enabledHint": "开启后，充值和兑换到账时会通过 Webhook 推送通知到飞书群",
  "webhookUrl": "Webhook URL",
  "webhookUrlPlaceholder": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
  "testSend": "测试发送",
  "testSendSuccess": "测试消息发送成功",
  "testSendFailed": "测试消息发送失败",
  "events": "推送事件",
  "rechargeEnabled": "直充到账推送",
  "rechargeEnabledHint": "用户通过支付宝/微信/Stripe 直充到账时推送",
  "redeemEnabled": "兑换到账推送",
  "redeemEnabledHint": "用户使用兑换码到账时推送",
  "fields": "显示字段",
  "fieldUserEmail": "用户邮箱",
  "fieldUserName": "用户名",
  "fieldUserID": "用户ID",
  "fieldAmount": "支付金额",
  "fieldCreditedAmount": "实际到账",
  "fieldMethod": "充值方式",
  "fieldMethodDetail": "支付渠道",
  "fieldOrderNo": "订单号/兑换码",
  "fieldTime": "时间"
}
```

- [ ] **Step 2: Add English i18n keys**

```json
"notification": {
  "title": "Notifications",
  "enabled": "Enable Feishu Push",
  "enabledHint": "When enabled, recharge and redeem notifications will be pushed to Feishu via webhook",
  "webhookUrl": "Webhook URL",
  "webhookUrlPlaceholder": "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
  "testSend": "Test Send",
  "testSendSuccess": "Test message sent successfully",
  "testSendFailed": "Test message failed to send",
  "events": "Push Events",
  "rechargeEnabled": "Direct Recharge Push",
  "rechargeEnabledHint": "Push when users recharge via Alipay/WeChat/Stripe",
  "redeemEnabled": "Redeem Code Push",
  "redeemEnabledHint": "Push when users redeem a code",
  "fields": "Display Fields",
  "fieldUserEmail": "User Email",
  "fieldUserName": "User Name",
  "fieldUserID": "User ID",
  "fieldAmount": "Payment Amount",
  "fieldCreditedAmount": "Credited Amount",
  "fieldMethod": "Recharge Method",
  "fieldMethodDetail": "Payment Channel",
  "fieldOrderNo": "Order No.",
  "fieldTime": "Time"
}
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/i18n/locales/zh-CN.json frontend/src/i18n/locales/en-US.json
git commit -m "feat: add feishu notification i18n keys"
```

---

### Task 15: Add feishu notification tab to SettingsView

**Files:**
- Modify: `frontend/src/views/admin/SettingsView.vue`

- [ ] **Step 1: Add `notification` to `SettingsTab` type and `settingsTabs` array**

After `"backup"` in the SettingsTab type:
```typescript
type SettingsTab =
  | "general"
  | "features"
  | "security"
  | "users"
  | "gateway"
  | "payment"
  | "email"
  | "notification"
  | "backup";
```

In `settingsTabs` array, add before backup:
```typescript
{ key: "notification" as SettingsTab, icon: "bell" as const },
```

- [ ] **Step 2: Add form fields**

In the `SettingsForm` type and the `form` reactive default, add:

```typescript
// In SettingsForm type (use Omit + extend pattern, these fields come from SystemSettings directly)
// In the form reactive initial value:
feishu_notify_enabled: false,
feishu_notify_webhook_url: "",
feishu_notify_recharge_enabled: true,
feishu_notify_redeem_enabled: true,
feishu_notify_fields: ["user_email", "user_name", "amount", "credited_amount", "method", "method_detail", "order_no", "time"],
```

- [ ] **Step 3: Add state variables**

```typescript
const testingFeishu = ref(false);
const feishuTestResult = ref<{ success: boolean; message: string } | null>(null);
const feishuFieldOptions = [
  { key: "user_email", labelZh: "用户邮箱", labelEn: "User Email" },
  { key: "user_name", labelZh: "用户名", labelEn: "User Name" },
  { key: "user_id", labelZh: "用户ID", labelEn: "User ID" },
  { key: "amount", labelZh: "支付金额", labelEn: "Payment Amount" },
  { key: "credited_amount", labelZh: "实际到账", labelEn: "Credited Amount" },
  { key: "method", labelZh: "充值方式", labelEn: "Recharge Method" },
  { key: "method_detail", labelZh: "支付渠道", labelEn: "Payment Channel" },
  { key: "order_no", labelZh: "订单号/兑换码", labelEn: "Order No." },
  { key: "time", labelZh: "时间", labelEn: "Time" },
];
```

- [ ] **Step 4: Add toggleField helper**

```typescript
function toggleFeishuField(key: string) {
  const idx = form.feishu_notify_fields.indexOf(key);
  if (idx >= 0) {
    form.feishu_notify_fields.splice(idx, 1);
  } else {
    form.feishu_notify_fields.push(key);
  }
}
```

- [ ] **Step 5: Add test send handler**

```typescript
async function testFeishuWebhook() {
  testingFeishu.value = true;
  feishuTestResult.value = null;
  try {
    await adminAPI.settings.testFeishu();
    feishuTestResult.value = { success: true, message: t("admin.settings.notification.testSendSuccess") };
  } catch (e: unknown) {
    feishuTestResult.value = {
      success: false,
      message: extractApiErrorMessage(e, t("admin.settings.notification.testSendFailed")),
    };
  } finally {
    testingFeishu.value = false;
  }
}
```

- [ ] **Step 6: Add tab template**

Add after the email tab div (`v-show="activeTab === 'email'"`) but before the backup tab:

```html
<div v-show="activeTab === 'notification'" class="space-y-6">
  <!-- Master toggle -->
  <div class="settings-card">
    <div class="settings-card-header">
      <h3>{{ t("admin.settings.notification.enabled") }}</h3>
      <p class="settings-card-hint">{{ t("admin.settings.notification.enabledHint") }}</p>
    </div>
    <div class="settings-card-body">
      <Toggle v-model="form.feishu_notify_enabled" />
    </div>
  </div>

  <!-- Webhook URL -->
  <div class="settings-card">
    <div class="settings-card-header">
      <h3>{{ t("admin.settings.notification.webhookUrl") }}</h3>
    </div>
    <div class="settings-card-body">
      <div class="flex gap-2">
        <input
          v-model="form.feishu_notify_webhook_url"
          type="url"
          class="form-input flex-1"
          :placeholder="t('admin.settings.notification.webhookUrlPlaceholder')"
        />
        <button
          class="btn btn-secondary"
          :disabled="testingFeishu || !form.feishu_notify_webhook_url"
          @click="testFeishuWebhook"
        >
          {{ testingFeishu ? "..." : t("admin.settings.notification.testSend") }}
        </button>
      </div>
      <p
        v-if="feishuTestResult"
        :class="feishuTestResult.success ? 'text-green-600' : 'text-red-600'"
        class="text-sm mt-2"
      >
        {{ feishuTestResult.message }}
      </p>
    </div>
  </div>

  <!-- Event toggles -->
  <div class="settings-card">
    <div class="settings-card-header">
      <h3>{{ t("admin.settings.notification.events") }}</h3>
    </div>
    <div class="settings-card-body space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <p>{{ t("admin.settings.notification.rechargeEnabled") }}</p>
          <p class="settings-card-hint">{{ t("admin.settings.notification.rechargeEnabledHint") }}</p>
        </div>
        <Toggle v-model="form.feishu_notify_recharge_enabled" />
      </div>
      <div class="flex items-center justify-between">
        <div>
          <p>{{ t("admin.settings.notification.redeemEnabled") }}</p>
          <p class="settings-card-hint">{{ t("admin.settings.notification.redeemEnabledHint") }}</p>
        </div>
        <Toggle v-model="form.feishu_notify_redeem_enabled" />
      </div>
    </div>
  </div>

  <!-- Field selection -->
  <div class="settings-card">
    <div class="settings-card-header">
      <h3>{{ t("admin.settings.notification.fields") }}</h3>
    </div>
    <div class="settings-card-body">
      <div class="grid grid-cols-3 gap-3">
        <label
          v-for="field in feishuFieldOptions"
          :key="field.key"
          class="flex items-center gap-2 cursor-pointer"
        >
          <input
            type="checkbox"
            :checked="form.feishu_notify_fields.includes(field.key)"
            class="checkbox"
            @change="toggleFeishuField(field.key)"
          />
          <span class="text-sm">{{ localText(field.labelZh, field.labelEn) }}</span>
        </label>
      </div>
    </div>
  </div>
</div>
```

- [ ] **Step 6: Add feishu fields to saveSettings payload**

In the `saveSettings` function, add to the `payload`:

```typescript
feishu_notify_enabled: form.feishu_notify_enabled,
feishu_notify_webhook_url: form.feishu_notify_webhook_url,
feishu_notify_recharge_enabled: form.feishu_notify_recharge_enabled,
feishu_notify_redeem_enabled: form.feishu_notify_redeem_enabled,
feishu_notify_fields: form.feishu_notify_fields,
```

- [ ] **Step 7: Import testFeishu API**

In the `<script setup>` imports, ensure `adminAPI.settings.testFeishu` is accessible. If `adminAPI` is imported via `@/api`, check the import. The testFeishu function should be exported from `@/api/admin/settings` and available as `adminAPI.settings.testFeishu`.

- [ ] **Step 8: Type check the frontend**

```bash
pnpm --dir frontend run typecheck
```

Expected: no errors

- [ ] **Step 9: Commit**

```bash
git add frontend/src/views/admin/SettingsView.vue
git commit -m "feat: add feishu notification tab to admin settings UI"
```

---

### Task 16: Verify, run tests, and final commit

- [ ] **Step 1: Full backend build**

```bash
cd backend && go build ./...
```

Expected: no errors

- [ ] **Step 2: Run backend tests**

```bash
cd backend && make test-unit
```

Expected: existing tests pass

- [ ] **Step 3: Frontend type check**

```bash
pnpm --dir frontend run typecheck
```

Expected: no errors

- [ ] **Step 4: Frontend lint**

```bash
pnpm --dir frontend run lint:check
```

Expected: no errors (or only pre-existing issues)

- [ ] **Step 5: Final commit if any fixes needed**

```bash
git add -A
git commit -m "chore: final adjustments for feishu notify feature"
```
