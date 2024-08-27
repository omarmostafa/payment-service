package requests

type WebhookEvent struct {
	EventType  string  `json:"eventType"`
	EventID    string  `json:"eventId"`
	MerchantID string  `json:"merchantId"`
	CreatedAt  string  `json:"eventDate"`
	Payload    Payload `json:"payload"`
}

type Payload struct {
	ID            string `json:"id"`
	ResponseCode  string `json:"responseCode"`
	AuthCode      string `json:"authCode"`
	TransactionID string `json:"transId"`
	AccountNumber string `json:"accountNumber"`
	AccountType   string `json:"accountType"`
}
