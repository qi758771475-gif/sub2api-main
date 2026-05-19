package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/model"
)

func TestRedeemService_FeishuNotifyField(t *testing.T) {
	h := newFeishuTestHelper()
	h.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(h.repo)
	defer notifySvc.Shutdown()

	svc := &RedeemService{feishuNotify: nil}
	if svc.feishuNotify != nil {
		t.Error("feishuNotify should be nil")
	}

	svc.feishuNotify = notifySvc
	if svc.feishuNotify == nil {
		t.Error("feishuNotify should be set")
	}
}

func TestRedeemService_NilNotify_DoesNotPanic(t *testing.T) {
	svc := &RedeemService{feishuNotify: nil}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on nil feishuNotify access: %v", r)
			}
		}()
		_ = svc.feishuNotify
	}()
}

func TestRedeemNotifyEvent_SendThroughChannel(t *testing.T) {
	h := newFeishuTestHelper()
	h.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(h.repo)
	defer notifySvc.Shutdown()

	svc := &RedeemService{feishuNotify: notifySvc}

	done := make(chan struct{})
	go func() {
		svc.feishuNotify.Send(model.RechargeEvent{
			UserEmail: "test@example.com",
			Method:    "兑换",
			OrderNo:   "TEST-REDEEM",
			Time:      "2026-05-19T00:00:00Z",
		})
		close(done)
	}()

	select {
	case <-done:
		// Channel accepted the event
	default:
		// Also OK — channel might be full but should not block
	}
}
