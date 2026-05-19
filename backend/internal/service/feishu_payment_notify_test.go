package service

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
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
	// PaymentService with nil feishuNotify should not panic
	svc := &PaymentService{feishuNotify: nil}
	o := &dbent.PaymentOrder{OutTradeNo: "test", Amount: 10.0}

	// This should return immediately without panicking
	svc.sendFeishuRechargeNotify(context.Background(), o)
}

func TestSendFeishuRechargeNotify_NilService_WithUserLookup(t *testing.T) {
	svc := &PaymentService{feishuNotify: nil}
	o := &dbent.PaymentOrder{OutTradeNo: "test", Amount: 10.0, UserID: 99999}

	// Should handle user not found gracefully
	svc.sendFeishuRechargeNotify(context.Background(), o)
}

func TestPaymentService_SetFeishuNotify(t *testing.T) {
	repo := newMockSettingRepo()
	repo.enableFeishu("http://localhost")
	notifySvc := NewFeishuNotifyService(repo)
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

func TestPaymentOrderFields_UsedByNotify(t *testing.T) {
	// Verify the fields used by sendFeishuRechargeNotify exist on PaymentOrder
	o := &dbent.PaymentOrder{
		OutTradeNo:  "sub2_test",
		Amount:      99.99,
		UserID:      1,
		PaymentType: "stripe",
	}

	if o.OutTradeNo != "sub2_test" {
		t.Error("OutTradeNo field should be accessible")
	}
	if o.Amount != 99.99 {
		t.Error("Amount field should be accessible")
	}
	if o.UserID != 1 {
		t.Error("UserID field should be accessible")
	}
	if o.PaymentType != "stripe" {
		t.Error("PaymentType field should be accessible")
	}

	// Verify that paymentorder column constants exist
	if paymentorder.FieldOutTradeNo == "" {
		t.Error("FieldOutTradeNo should not be empty")
	}
}
