/AccountQuotaNotifyEmails.*json:"account_quota_notify_emails"/{
n
a\
\t\t// Feishu webhook notification\
\t\tFeishuNotifyEnabled         bool     `json:"feishu_notify_enabled"`\
\t\tFeishuNotifyWebhookURL      string   `json:"feishu_notify_webhook_url"`\
\t\tFeishuNotifyRechargeEnabled bool     `json:"feishu_notify_recharge_enabled"`\
\t\tFeishuNotifyRedeemEnabled   bool     `json:"feishu_notify_redeem_enabled"`\
\t\tFeishuNotifyFields          []string `json:"feishu_notify_fields"`
}
