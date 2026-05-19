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
	shutdown    chan struct{}

	mu              sync.RWMutex
	enabled         bool
	webhookURL      string
	rechargeEnabled bool
	redeemEnabled   bool
	visibleFields   []string
	configExpiresAt time.Time
}

// NewFeishuNotifyService creates the service and starts the background worker.
func NewFeishuNotifyService(settingRepo SettingRepository) *FeishuNotifyService {
	s := &FeishuNotifyService{
		ch:          make(chan model.RechargeEvent, 64),
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		settingRepo: settingRepo,
		shutdown:    make(chan struct{}),
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

// Shutdown gracefully stops the worker, draining any remaining events before returning.
func (s *FeishuNotifyService) Shutdown() {
	close(s.shutdown)
}

func (s *FeishuNotifyService) worker() {
	for {
		select {
		case <-s.shutdown:
			// Drain remaining events before exiting.
			for {
				select {
				case event, ok := <-s.ch:
					if !ok {
						return
					}
					s.processEvent(event)
				default:
					return
				}
			}
		case event, ok := <-s.ch:
			if !ok {
				return
			}
			s.processEvent(event)
		}
	}
}

func (s *FeishuNotifyService) processEvent(event model.RechargeEvent) {
	ctx := context.Background()
	s.reloadConfig(ctx)

	s.mu.RLock()
	enabled := s.enabled
	rechargeEnabled := s.rechargeEnabled
	redeemEnabled := s.redeemEnabled
	webhookURL := s.webhookURL
	s.mu.RUnlock()

	if !enabled {
		return
	}
	if event.Method == "直冲" && !rechargeEnabled {
		return
	}
	if event.Method == "兑换" && !redeemEnabled {
		return
	}
	if webhookURL == "" {
		return
	}
	card := s.buildCard(event)
	if err := s.httpPost(ctx, webhookURL, card); err != nil {
		slog.Warn("feishu notify send failed", "error", err)
	}
}

func (s *FeishuNotifyService) reloadConfig(ctx context.Context) {
	s.mu.RLock()
	if time.Now().Before(s.configExpiresAt) {
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()

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
	s.configExpiresAt = time.Now().Add(30 * time.Second)
}

func (s *FeishuNotifyService) buildCard(event model.RechargeEvent) map[string]any {
	headerTitle := "充值到账通知"
	if event.Method == "兑换" {
		headerTitle = "兑换到账通知"
	}

	s.mu.RLock()
	visibleFields := s.visibleFields
	s.mu.RUnlock()

	fieldSet := make(map[string]bool, len(visibleFields))
	for _, f := range visibleFields {
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

func (s *FeishuNotifyService) httpPost(ctx context.Context, url string, card map[string]any) error {
	body, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("marshal card: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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

// SendTestCard sends a sample notification card using the configured webhook URL.
func (s *FeishuNotifyService) SendTestCard(ctx context.Context) error {
	s.reloadConfig(ctx)
	s.mu.RLock()
	url := s.webhookURL
	s.mu.RUnlock()
	return s.sendTestCard(ctx, url)
}

// SendTestCardWithURL sends a sample notification card to the given webhook URL.
// The frontend test-send button provides the URL from the form input without saving first.
func (s *FeishuNotifyService) SendTestCardWithURL(ctx context.Context, url string) error {
	s.reloadConfig(ctx)
	return s.sendTestCard(ctx, url)
}

func (s *FeishuNotifyService) sendTestCard(ctx context.Context, url string) error {
	if url == "" {
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
	return s.httpPost(ctx, url, card)
}
