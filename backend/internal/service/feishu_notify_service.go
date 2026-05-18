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
