package providers

import (
	"encoding/json"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/paymentintent"
	"github.com/stripe/stripe-go/payout"
	"payment-service/app"
	"payment-service/domain/entities"
	"payment-service/domain/types"
	"payment-service/errors"
	"time"
)

type StripePaymentProvider struct {
}

func NewStripePaymentProvider() *StripePaymentProvider {
	return &StripePaymentProvider{}
}

func (self *StripePaymentProvider) Charge(params types.DepositParams, transaction entities.Transaction) (entities.Transaction, error) {
	stripe.Key = app.App().Config().GetString("payment.stripe_secret_key")
	stripeParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(params.Amount)),
		Currency: stripe.String(params.Currency),
		Confirm:  stripe.Bool(true),
	}
	stripeParams.SetIdempotencyKey(params.TransactionId)

	maskedRequest := stripeParams
	stripeParamsJson, _ := json.Marshal(maskedRequest)
	stripeParamsStr := string(stripeParamsJson)
	transaction.RequestPayload = &stripeParamsStr

	stripeParams.PaymentMethod = stripe.String(params.Token)

	var paymentIntent *stripe.PaymentIntent
	var err error
	attempts := 3
	sleep := 2 * time.Second

	for i := 0; i < attempts; i++ {
		paymentIntent, err = paymentintent.New(stripeParams)
		if err == nil {
			break
		}
		if stripeErr, ok := err.(*stripe.Error); ok {
			if !isRetryable(stripeErr) {
				break
			}
		} else {
			break
		}

		time.Sleep(sleep)
		sleep = sleep * 2
	}
	if err != nil {
		app.App().Logger().Error("failed to create payment intent: ", err.Error())
		var paymentError types.PaymentIntentError
		json.Unmarshal([]byte(err.Error()), &paymentError)
		transaction.ChargeId = paymentError.Charge
		transaction.PaymentId = paymentError.PaymentIntent.ID
		return transaction, &errors.ValidationError{
			Message: paymentError.Message,
		}
	}
	transaction.PaymentId = paymentIntent.ID
	paymentIntentJson, _ := json.Marshal(paymentIntent)
	paymentIntentStr := string(paymentIntentJson)
	transaction.ResponsePayload = &paymentIntentStr

	app.App().Logger().Info("payment intent created: ", paymentIntent)

	return transaction, nil
}

func (self *StripePaymentProvider) Withdraw(params types.WithdrawParams, transaction entities.Transaction) (entities.Transaction, error) {
	stripe.Key = app.App().Config().GetString("payment.stripe_secret_key")
	payoutParams := &stripe.PayoutParams{
		Amount:      stripe.Int64(params.Amount),
		Currency:    stripe.String(params.Currency),
		Destination: stripe.String(params.Destination),
	}
	payoutParams.SetIdempotencyKey(params.TransactionId)

	maskedRequest := payoutParams
	stripeParamsJson, _ := json.Marshal(maskedRequest)
	stripeParamsStr := string(stripeParamsJson)
	transaction.RequestPayload = &stripeParamsStr

	payoutParams.Destination = stripe.String(params.Destination)

	payout, err := payout.New(payoutParams)

	if err != nil {
		app.App().Logger().Error("failed to create payout: ", err.Error())
		var paymentError types.PaymentIntentError
		json.Unmarshal([]byte(err.Error()), &paymentError)
		transaction.ChargeId = paymentError.Charge
		transaction.PaymentId = paymentError.PaymentIntent.ID
		return transaction, &errors.ValidationError{
			Message: paymentError.Message,
		}
	}
	transaction.PaymentId = payout.ID
	payoutJson, _ := json.Marshal(payout)
	payoutStr := string(payoutJson)
	transaction.ResponsePayload = &payoutStr

	app.App().Logger().Info("payout created: ", payoutStr)

	return transaction, nil
}

func isRetryable(err *stripe.Error) bool {
	switch err.Code {
	case stripe.ErrorCodeRateLimit, stripe.ErrorCodeLockTimeout:
		return true
	default:
		return false
	}
}
