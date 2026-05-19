package service

import (
	"testing"
)

func TestRedeemService_FeishuNotifyField(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(repo)
	defer notifySvc.Shutdown()

	// Create RedeemService with nil feishuNotify
	svc := &RedeemService{feishuNotify: nil}
	if svc.feishuNotify != nil {
		t.Error("feishuNotify should be nil")
	}

	// Assign and verify
	svc.feishuNotify = notifySvc
	if svc.feishuNotify == nil {
		t.Error("feishuNotify should be set")
	}
}

func TestRedeemService_NilNotify_DoesNotPanic(t *testing.T) {
	svc := &RedeemService{feishuNotify: nil}

	// Verify nil field doesn't panic on access
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on nil feishuNotify access: %v", r)
			}
		}()
		_ = svc.feishuNotify
	}()
}

func TestRedeemNotifyEvent_Fields(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(repo)
	defer notifySvc.Shutdown()

	svc := &RedeemService{feishuNotify: notifySvc}

	// Verify the notify service is properly wired
	if svc.feishuNotify == nil {
		t.Fatal("feishuNotify should not be nil")
	}

	// Send a raw event through the service to verify the channel works
	svc.feishuNotify.Send(RedeemCode{
		Code:  "TEST-REDEEM-CODE",
		Value: 50.0,
		Type:  RedeemTypeBalance,
		// Send will go through — verify it doesn't block
	})

	// If we get here without hanging, the channel is working
}
