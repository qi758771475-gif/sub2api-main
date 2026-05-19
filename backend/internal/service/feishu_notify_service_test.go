package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/model"
)

// mockSettingRepo implements SettingRepository for tests.
type mockSettingRepo struct {
	mu   sync.RWMutex
	data map[string]string
}

func newMockSettingRepo() *mockSettingRepo {
	return &mockSettingRepo{data: make(map[string]string)}
}

func (m *mockSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if v, ok := m.data[key]; ok {
		return &Setting{Key: key, Value: v}, nil
	}
	return nil, nil
}

func (m *mockSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[key], nil
}

func (m *mockSettingRepo) Set(ctx context.Context, key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]string, len(keys))
	for _, k := range keys {
		result[k] = m.data[k]
	}
	return result, nil
}

func (m *mockSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range settings {
		m.data[k] = v
	}
	return nil
}

func (m *mockSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]string, len(m.data))
	for k, v := range m.data {
		result[k] = v
	}
	return result, nil
}

func (m *mockSettingRepo) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// enableFeishu sets the mock repo to have feishu enabled with a given webhook URL.
func (m *mockSettingRepo) enableFeishu(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[SettingKeyFeishuNotifyEnabled] = "true"
	m.data[SettingKeyFeishuNotifyWebhookURL] = url
	m.data[SettingKeyFeishuNotifyRechargeEnabled] = "true"
	m.data[SettingKeyFeishuNotifyRedeemEnabled] = "true"
	fields, _ := json.Marshal(model.DefaultFeishuNotifyFields())
	m.data[SettingKeyFeishuNotifyFields] = string(fields)
}

// --- buildCard tests ---

func TestBuildCard_AllFields_Recharge(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond) // wait for worker goroutine

	svc.reloadConfig(context.Background())
	svc.visibleFields = model.DefaultFeishuNotifyFields()

	event := model.RechargeEvent{
		UserEmail:      "test@example.com",
		UserName:       "testuser",
		UserID:         42,
		Amount:         100.50,
		CreditedAmount: 100.50,
		Method:         "直冲",
		MethodDetail:   "支付宝",
		OrderNo:        "sub2_test123",
		Time:           "2026-05-19T12:00:00Z",
	}

	card := svc.buildCard(event)

	// Verify header
	if card["msg_type"] != "interactive" {
		t.Errorf("msg_type = %v, want interactive", card["msg_type"])
	}
	cardMap, ok := card["card"].(map[string]any)
	if !ok {
		t.Fatal("card field missing or wrong type")
	}
	header, ok := cardMap["header"].(map[string]any)
	if !ok {
		t.Fatal("header missing")
	}
	title, ok := header["title"].(map[string]string)
	if !ok {
		t.Fatal("title missing")
	}
	if title["content"] != "充值到账通知" {
		t.Errorf("title = %s, want 充值到账通知", title["content"])
	}
	if header["template"] != "blue" {
		t.Errorf("template = %v, want blue", header["template"])
	}

	// Verify elements
	elements, ok := cardMap["elements"].([]map[string]any)
	if !ok {
		t.Fatal("elements missing or wrong type")
	}
	if len(elements) != 8 {
		t.Errorf("elements count = %d, want 8 (all default fields)", len(elements))
	}
}

func TestBuildCard_RedeemMethod_ChangesTitle(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond)

	svc.reloadConfig(context.Background())
	svc.visibleFields = model.DefaultFeishuNotifyFields()

	event := model.RechargeEvent{
		UserEmail: "test@example.com",
		Method:    "兑换",
	}
	card := svc.buildCard(event)
	cardMap := card["card"].(map[string]any)
	header := cardMap["header"].(map[string]any)
	title := header["title"].(map[string]string)
	if title["content"] != "兑换到账通知" {
		t.Errorf("title = %s, want 兑换到账通知", title["content"])
	}
}

func TestBuildCard_FieldFiltering(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond)

	svc.reloadConfig(context.Background())
	// Only show user_email and amount
	svc.visibleFields = []string{model.FeishuFieldUserEmail, model.FeishuFieldAmount}

	event := model.RechargeEvent{
		UserEmail: "test@example.com",
		Amount:    50.0,
		Method:    "直冲",
	}
	card := svc.buildCard(event)
	cardMap := card["card"].(map[string]any)
	elements := cardMap["elements"].([]map[string]any)
	if len(elements) != 2 {
		t.Errorf("elements count = %d, want 2 (only user_email and amount)", len(elements))
	}
}

func TestBuildCard_EmptyFields(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond)

	svc.reloadConfig(context.Background())
	svc.visibleFields = []string{}

	event := model.RechargeEvent{Method: "直冲"}
	card := svc.buildCard(event)
	cardMap := card["card"].(map[string]any)
	elements := cardMap["elements"].([]map[string]any)
	if len(elements) != 0 {
		t.Errorf("elements count = %d, want 0", len(elements))
	}
}

// --- Send tests ---

func TestSend_NonBlocking(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	// Sending should not block even with a small buffer
	for i := 0; i < 100; i++ {
		svc.Send(model.RechargeEvent{Method: "直冲"})
	}
	// If it doesn't hang, test passes
}

// --- Shutdown tests ---

func TestShutdown_DrainsRemainingEvents(t *testing.T) {
	received := make(chan struct{}, 10)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		received <- struct{}{}
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	repo.enableFeishu(srv.URL)
	svc := NewFeishuNotifyService(repo)

	// Send an event then shutdown immediately
	svc.Send(model.RechargeEvent{
		UserEmail: "test@example.com",
		Method:    "直冲",
		Time:      time.Now().Format(time.RFC3339),
	})

	svc.Shutdown()

	// Should have received the event (drained before exit)
	select {
	case <-received:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("shutdown did not drain remaining events")
	}
}

// --- reloadConfig tests ---

func TestReloadConfig_LoadsAllSettings(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("https://open.feishu.cn/hook/test")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	svc.reloadConfig(context.Background())

	if !svc.enabled {
		t.Error("enabled should be true")
	}
	if svc.webhookURL != "https://open.feishu.cn/hook/test" {
		t.Errorf("webhookURL = %s", svc.webhookURL)
	}
	if !svc.rechargeEnabled {
		t.Error("rechargeEnabled should be true")
	}
	if !svc.redeemEnabled {
		t.Error("redeemEnabled should be true")
	}
	if len(svc.visibleFields) != 8 {
		t.Errorf("visibleFields count = %d, want 8", len(svc.visibleFields))
	}
}

func TestReloadConfig_DisabledByDefault(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	svc.reloadConfig(context.Background())

	if svc.enabled {
		t.Error("enabled should be false when key is missing")
	}
}

func TestReloadConfig_CacheTTL(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://first-url.local")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	svc.reloadConfig(context.Background())
	if svc.webhookURL != "http://first-url.local" {
		t.Fatalf("first load failed: %s", svc.webhookURL)
	}

	// Change the URL in the repo without calling reloadConfig
	repo.Set(context.Background(), SettingKeyFeishuNotifyWebhookURL, "http://second-url.local")

	// reloadConfig should return cached version (TTL not expired)
	svc.reloadConfig(context.Background())
	if svc.webhookURL != "http://first-url.local" {
		t.Error("config cache should have prevented reload, but URL changed")
	}

	// Force cache expiry
	svc.configExpiresAt = time.Time{}
	svc.reloadConfig(context.Background())
	if svc.webhookURL != "http://second-url.local" {
		t.Error("after cache expiry, URL should have updated")
	}
}

func TestReloadConfig_DefaultFieldsWhenEmpty(t *testing.T) {
	repo := newMockSettingRepo()
	repo.data = map[string]string{
		SettingKeyFeishuNotifyEnabled:    "true",
		SettingKeyFeishuNotifyWebhookURL: "http://localhost",
	}
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	svc.reloadConfig(context.Background())

	if len(svc.visibleFields) != len(model.DefaultFeishuNotifyFields()) {
		t.Errorf("visibleFields should fallback to defaults, got %d fields", len(svc.visibleFields))
	}
}

// --- SendTestCard / SendTestCardWithURL tests ---

func TestSendTestCard_SendsToConfiguredURL(t *testing.T) {
	received := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		var body []byte
		r.Body.Read(body)
		received <- body
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	repo.enableFeishu(srv.URL)
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	err := svc.SendTestCard(context.Background())
	if err != nil {
		t.Errorf("SendTestCard returned error: %v", err)
	}
}

func TestSendTestCard_EmptyURL_ReturnsError(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	err := svc.SendTestCard(context.Background())
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestSendTestCardWithURL_UsesProvidedURL(t *testing.T) {
	received := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		received <- struct{}{}
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	// Don't set a URL in repo — use the URL parameter instead
	repo.enableFeishu("")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	err := svc.SendTestCardWithURL(context.Background(), srv.URL)
	if err != nil {
		t.Errorf("SendTestCardWithURL returned error: %v", err)
	}

	select {
	case <-received:
		// OK
	case <-time.After(1 * time.Second):
		t.Error("server did not receive request")
	}
}

func TestSendTestCardWithURL_EmptyURL_ReturnsError(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	err := svc.SendTestCardWithURL(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

// --- HTTP error handling tests ---

func TestHTTPPost_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	repo.enableFeishu(srv.URL)
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	card := map[string]any{"msg_type": "interactive"}
	err := svc.httpPost(context.Background(), srv.URL, card)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestHTTPPost_InvalidURL(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()

	card := map[string]any{"msg_type": "interactive"}
	err := svc.httpPost(context.Background(), "http://127.0.0.1:1/nothing", card)
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

// --- Integration: processEvent filtering ---

func TestProcessEvent_RespectsEnabledFlags(t *testing.T) {
	received := make(chan struct{}, 5)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		received <- struct{}{}
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	repo.enableFeishu(srv.URL)
	// Disable recharge events
	repo.Set(context.Background(), SettingKeyFeishuNotifyRechargeEnabled, "false")

	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond)

	// Force reload to pick up disabled recharge
	svc.configExpiresAt = time.Time{}

	// Send a recharge event — should be filtered out
	svc.Send(model.RechargeEvent{Method: "直冲", Time: time.Now().Format(time.RFC3339)})
	time.Sleep(100 * time.Millisecond)

	select {
	case <-received:
		t.Error("recharge event should not have been sent (disabled)")
	default:
		// OK
	}
}

func TestProcessEvent_SendsRedeemWhenEnabled(t *testing.T) {
	received := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		received <- struct{}{}
	}))
	defer srv.Close()

	repo := newMockSettingRepo()
	repo.enableFeishu(srv.URL)
	svc := NewFeishuNotifyService(repo)
	defer svc.Shutdown()
	time.Sleep(10 * time.Millisecond)

	svc.Send(model.RechargeEvent{
		UserEmail: "test@example.com",
		Method:    "兑换",
		Time:      time.Now().Format(time.RFC3339),
	})

	select {
	case <-received:
		// OK
	case <-time.After(2 * time.Second):
		t.Error("redeem event should have been sent")
	}
}
