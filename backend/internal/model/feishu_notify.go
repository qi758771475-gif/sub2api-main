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
