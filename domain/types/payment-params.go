package types

import "github.com/stripe/stripe-go"

type DepositParams struct {
	Amount           float64
	Currency         string
	Token            string
	TransactionId    string
	UserId           string
	Provider         string
	CreditCardNumber string
	ExpirationDate   string
	CVV              string
}

type WithdrawParams struct {
	Amount           int64
	Currency         string
	Destination      string
	TransactionId    string
	UserId           string
	Provider         string
	CreditCardNumber string
	ExpirationDate   string
	CVV              string
}

type CustomPaymentIntent struct {
	LatestCharge string `json:"latest_charge"`
}

type PaymentIntentError struct {
	Message       string               `json:"message"`
	Code          string               `json:"code"`
	Charge        string               `json:"charge"`
	PaymentIntent stripe.PaymentIntent `json:"payment_intent"`
}
