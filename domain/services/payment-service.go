package services

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/refund"
	"net/http"
	"payment-service/app"
	"payment-service/domain/entities"
	"payment-service/domain/repositories"
	"payment-service/domain/types"
	"payment-service/errors"
	"payment-service/interfaces"
	"payment-service/requests"
	"strings"
)

type PaymentService struct {
	PaymentProvider       interfaces.IPaymentProvider
	TransactionRepository *repositories.TransactionRepository
}

func NewPaymentService() *PaymentService {
	return &PaymentService{
		TransactionRepository: repositories.NewTransactionRepository(),
	}
}

func (self *PaymentService) SetPaymentProvider(provider interfaces.IPaymentProvider) {
	self.PaymentProvider = provider
}

func (self *PaymentService) Deposit(params types.DepositParams) (*entities.Transaction, error) {
	if self.PaymentProvider == nil {
		return nil, &errors.ValidationError{
			Message: "Payment provider is not set",
		}
	}

	existingTransaction, err := self.TransactionRepository.GetTransactionByTransactionId(params.TransactionId, nil)
	if err != nil {
		return nil, &errors.InternalServerError{
			Message: err.Error(),
		}
	}

	if existingTransaction != nil {
		app.App().Logger().Error("transaction already exists: ", params.TransactionId)
		return nil, &errors.ValidationError{
			Message: "Transaction already exists",
		}
	}

	transaction, err := self.TransactionRepository.SaveTransaction(entities.Transaction{
		Amount:          params.Amount,
		Currency:        params.Currency,
		TransactionID:   params.TransactionId,
		Status:          "pending",
		TransactionType: "deposit",
		GatewayName:     params.Provider,
	}, nil)

	if err != nil {
		app.App().Logger().Error("failed to save transaction in initial state: ", err.Error())
		return &transaction, &errors.InternalServerError{
			Message: err.Error(),
		}
	}

	transaction, err = self.PaymentProvider.Charge(params, transaction)
	_, txErr := self.TransactionRepository.SaveTransaction(transaction, nil)
	if txErr != nil {
		app.App().Logger().Error("failed to save transaction after payment failed: ", txErr.Error())
		return nil, &errors.InternalServerError{
			Message: txErr.Error(),
		}
	}
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (self *PaymentService) Withdraw(params types.WithdrawParams) (*entities.Transaction, error) {
	if self.PaymentProvider == nil {
		return nil, &errors.ValidationError{
			Message: "Payment provider is not set",
		}
	}

	existingTransaction, err := self.TransactionRepository.GetTransactionByTransactionId(params.TransactionId, nil)
	if err != nil {
		return nil, &errors.InternalServerError{
			Message: err.Error(),
		}
	}

	if existingTransaction != nil {
		app.App().Logger().Error("transaction already exists: ", params.TransactionId)
		return nil, &errors.ValidationError{
			Message: "Transaction already exists",
		}
	}

	transaction, err := self.TransactionRepository.SaveTransaction(entities.Transaction{
		Amount:          float64(params.Amount),
		Currency:        params.Currency,
		TransactionID:   params.TransactionId,
		Status:          "pending",
		TransactionType: "withdrawal",
		GatewayName:     params.Provider,
	}, nil)

	if err != nil {
		app.App().Logger().Error("failed to save transaction in initial state: ", err.Error())
		return &transaction, &errors.InternalServerError{
			Message: err.Error(),
		}
	}

	transaction, err = self.PaymentProvider.Withdraw(params, transaction)
	_, txErr := self.TransactionRepository.SaveTransaction(transaction, nil)
	if txErr != nil {
		app.App().Logger().Error("failed to save transaction after payout failed: ", txErr.Error())
		return nil, &errors.InternalServerError{
			Message: txErr.Error(),
		}
	}
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (self *PaymentService) HandleStripeEvents(event stripe.Event) error {
	switch event.Type {
	case "payment_intent.succeeded":
		var paymentIntent stripe.PaymentIntent
		json.Unmarshal(event.Data.Raw, &paymentIntent)

		app.App().Logger().Info("New Payment Intent succeeded , Payment id ", paymentIntent.ID)

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(paymentIntent.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payment id: ", err.Error())
			return err
		}
		transaction.Status = "succeeded"

		var latestCharge types.CustomPaymentIntent
		json.Unmarshal(event.Data.Raw, &latestCharge)
		transaction.ChargeId = latestCharge.LatestCharge

		responsePayloadStr := ""
		transaction.ResponsePayload = &responsePayloadStr

		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)
		if err != nil {
			app.App().Logger().Error("failed to save transaction after payment intent Success: ", err.Error())

			// If saving the transaction fails, attempt to refund the payment
			refundParams := &stripe.RefundParams{
				Charge: stripe.String(latestCharge.LatestCharge),
			}
			_, refundErr := refund.New(refundParams)
			if refundErr != nil {
				app.App().Logger().Error("failed to refund the charge: ", refundErr.Error())
				return &errors.InternalServerError{
					Message: "Failed to save transaction and refund charge: " + err.Error() + " and " + refundErr.Error(),
				}
			}

			return &errors.InternalServerError{
				Message: "Failed to save transaction, but refund was successful: " + err.Error(),
			}
		}

		app.App().Logger().Info("Payment Succeeded, Payment id", paymentIntent.ID)
		return nil
	case "payment_intent.payment_failed":
		var paymentIntent stripe.PaymentIntent
		json.Unmarshal(event.Data.Raw, &paymentIntent)
		app.App().Logger().Info("New Payment Failed succeeded , Payment id", paymentIntent.ID)
		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(paymentIntent.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payment id: ", err.Error())
			return err
		}
		transaction.Status = "failed"
		responsePayloadStr := string(event.Data.Raw)
		transaction.ResponsePayload = &responsePayloadStr
		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)
		if err != nil {
			app.App().Logger().Error("failed to save transaction after payment intent Success: ", err.Error())
			return &errors.InternalServerError{
				Message: err.Error(),
			}
		}
		app.App().Logger().Info("Payment Failed, Payment id", paymentIntent.ID)
		return nil
	case "charge.refunded":
		var charge stripe.Charge
		json.Unmarshal(event.Data.Raw, &charge)
		app.App().Logger().Info("New Charge refunded , Payment id ", charge.PaymentIntent)

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(charge.PaymentIntent, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payment id: ", err.Error())
			return err
		}
		transaction.Status = "refunded"
		transaction.ChargeId = charge.ID
		responsePayloadStr := string(event.Data.Raw)
		transaction.ResponsePayload = &responsePayloadStr

		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)
		if err != nil {
			return &errors.InternalServerError{
				Message: "Failed to save refunded transaction" + err.Error(),
			}
		}

		app.App().Logger().Info("Charge Refunded, Payment id", charge.PaymentIntent)
		return nil
	case "payout.paid":
		var payout stripe.Payout
		json.Unmarshal(event.Data.Raw, &payout)
		app.App().Logger().Info("New Payout created , Payout id ", payout.ID)

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(payout.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payout id: ", err.Error())
			return err
		}
		transaction.Status = "succeeded"
		responsePayloadStr := string(event.Data.Raw)
		transaction.ResponsePayload = &responsePayloadStr

		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)
		if err != nil {
			return &errors.InternalServerError{
				Message: "Failed to save refunded transaction" + err.Error(),
			}
		}

		app.App().Logger().Info("Payout Paid", payout.ID)
		return nil
	case "payout.failed":
		var payout stripe.Payout
		json.Unmarshal(event.Data.Raw, &payout)
		app.App().Logger().Info("New Payout failed , Payout id ", payout.ID)

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(payout.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payout id: ", err.Error())
			return err
		}
		transaction.Status = "failed"
		responsePayloadStr := string(event.Data.Raw)
		transaction.ResponsePayload = &responsePayloadStr

		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)
		if err != nil {
			return &errors.InternalServerError{
				Message: "Failed to save refunded transaction" + err.Error(),
			}
		}

		app.App().Logger().Info("Payout Failed", payout.ID)
		return nil
	default:
		app.App().Logger().Info("Unhandled event type: ", event.Type)
	}
	return nil
}

func (self *PaymentService) HandleAuthorizeEvents(event requests.WebhookEvent) error {
	switch event.EventType {
	case "net.authorize.payment.authcapture.created":

		app.App().Logger().Info("New net.authorize.payment.authcapture.created ")

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(event.Payload.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payment id: ", err.Error())
			return err
		}
		transaction.Status = "succeeded"
		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)

		app.App().Logger().Info("Payment Succeeded, Payment id", event.Payload.ID)
		return nil
	case "net.authorize.payment.refund.created":

		app.App().Logger().Info("net.authorize.payment.refund.created ")

		transaction, err := self.TransactionRepository.GetTransactionByPaymentId(event.Payload.ID, nil)
		if err != nil {
			app.App().Logger().Error("failed to get transaction by payment id: ", err.Error())
			return err
		}
		transaction.Status = "succeeded"
		_, err = self.TransactionRepository.SaveTransaction(*transaction, nil)

		app.App().Logger().Info("Refund Succeeded, Payment id", event.Payload.ID)
		return nil
	default:
		app.App().Logger().Info("Unhandled event type: ", event.EventType)
	}
	return nil
}

func (self *PaymentService) VerifyAuthorizeSignature(r *http.Request, payload []byte) bool {
	signatureHeader := r.Header.Get("x-anet-signature")
	if signatureHeader == "" {
		return false
	}

	// Your webhook signature key from Authorize.Net
	webhookSignatureKey := app.App().Config().GetString("payment.authorize_net_webhook_signature_key")

	// Create HMAC with SHA512
	h := hmac.New(sha512.New, []byte(webhookSignatureKey))
	h.Write(payload)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare the expected signature with the signature in the header
	return strings.EqualFold(signatureHeader, "sha512="+expectedSignature)
}
