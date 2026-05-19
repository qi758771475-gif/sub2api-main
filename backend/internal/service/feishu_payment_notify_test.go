package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

func TestResolveProviderDisplayName(t *testing.T) {
	svc := &PaymentService{}

	tests := []struct {
		paymentType string
		want        string
	}{
		{"alipay", "支付宝"},
		{"Alipay", "支付宝"},
		{"ALIPAY", "支付宝"},
		{"wxpay", "微信支付"},
		{"wechat", "微信支付"},
		{"stripe", "Stripe"},
		{"easypay", "易支付"},
		{"unknown_provider", "unknown_provider"},
	}

	for _, tt := range tests {
		o := &dbent.PaymentOrder{PaymentType: tt.paymentType}
		got := svc.resolveProviderDisplayName(o)
		if got != tt.want {
			t.Errorf("resolveProviderDisplayName(%q) = %q, want %q", tt.paymentType, got, tt.want)
		}
	}
}

func TestSendFeishuRechargeNotify_NilService(t *testing.T) {
	svc := &PaymentService{feishuNotify: nil}
	o := &dbent.PaymentOrder{OutTradeNo: "test", Amount: 10.0}

	// Should return immediately without panicking
	svc.sendFeishuRechargeNotify(context.Background(), o)
}

func TestSendFeishuRechargeNotify_NilService_WithUserLookup(t *testing.T) {
	svc := &PaymentService{feishuNotify: nil}
	o := &dbent.PaymentOrder{OutTradeNo: "test", Amount: 10.0, UserID: 99999}

	// Should handle user not found gracefully
	svc.sendFeishuRechargeNotify(context.Background(), o)
}

func TestPaymentService_SetFeishuNotify(t *testing.T) {
	h := newFeishuTestHelper()
	h.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(h.repo)
	defer notifySvc.Shutdown()

	paymentSvc := &PaymentService{}
	if paymentSvc.feishuNotify != nil {
		t.Error("feishuNotify should be nil before SetFeishuNotify")
	}

	paymentSvc.SetFeishuNotify(notifySvc)
	if paymentSvc.feishuNotify == nil {
		t.Error("feishuNotify should not be nil after SetFeishuNotify")
	}
}
