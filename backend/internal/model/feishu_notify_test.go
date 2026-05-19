package model

import "testing"

func TestDefaultFeishuNotifyFields(t *testing.T) {
	fields := DefaultFeishuNotifyFields()

	if len(fields) == 0 {
		t.Fatal("DefaultFeishuNotifyFields returned empty slice")
	}

	// user_id should NOT be in defaults
	for _, f := range fields {
		if f == FeishuFieldUserID {
			t.Errorf("FeishuFieldUserID should not be in default fields")
		}
	}

	// All other 8 fields should be present
	expected := map[string]bool{
		FeishuFieldUserEmail:      false,
		FeishuFieldUserName:       false,
		FeishuFieldAmount:         false,
		FeishuFieldCreditedAmount: false,
		FeishuFieldMethod:         false,
		FeishuFieldMethodDetail:   false,
		FeishuFieldOrderNo:        false,
		FeishuFieldTime:           false,
	}
	for _, f := range fields {
		if _, ok := expected[f]; ok {
			expected[f] = true
		}
	}
	for k, v := range expected {
		if !v {
			t.Errorf("missing field %s in defaults", k)
		}
	}
}

func TestFeishuFieldConstants(t *testing.T) {
	// Verify constant values match expected field keys
	tests := []struct {
		constant string
		want     string
	}{
		{FeishuFieldUserEmail, "user_email"},
		{FeishuFieldUserName, "user_name"},
		{FeishuFieldUserID, "user_id"},
		{FeishuFieldAmount, "amount"},
		{FeishuFieldCreditedAmount, "credited_amount"},
		{FeishuFieldMethod, "method"},
		{FeishuFieldMethodDetail, "method_detail"},
		{FeishuFieldOrderNo, "order_no"},
		{FeishuFieldTime, "time"},
	}
	for _, tt := range tests {
		if tt.constant != tt.want {
			t.Errorf("constant = %s, want %s", tt.constant, tt.want)
		}
	}
}

func TestRechargeEventFields(t *testing.T) {
	event := RechargeEvent{
		UserID:         1,
		UserEmail:      "test@example.com",
		UserName:       "testuser",
		Amount:         100.00,
		CreditedAmount: 100.00,
		Method:         "直冲",
		MethodDetail:   "支付宝",
		OrderNo:        "TEST-001",
		Time:           "2026-05-19T00:00:00Z",
	}

	if event.Method != "直冲" {
		t.Errorf("Method = %s, want 直冲", event.Method)
	}
	if event.Amount != 100.00 {
		t.Errorf("Amount = %f, want 100.00", event.Amount)
	}
}
