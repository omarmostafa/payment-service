package requests

type DepositRequest struct {
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	UserId           string  `json:"userId" validate:"required"`
	Token            string  `json:"token"`
	Currency         string  `json:"currency" validate:"required,len=3"`
	TransactionId    string  `json:"transactionId" validate:"required"`
	Provider         string  `json:"provider" validate:"required,oneof=stripe authorize"`
	CreditCardNumber string  `json:"creditCardNumber"`
	ExpirationDate   string  `json:"expirationDate"`
	CVV              string  `json:"cvv"`
}
