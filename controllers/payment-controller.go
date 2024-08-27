package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/stripe/stripe-go/webhook"
	"io/ioutil"
	"net/http"
	"payment-service/app"
	"payment-service/domain/providers"
	"payment-service/domain/services"
	"payment-service/domain/types"
	"payment-service/errors"
	"payment-service/requests"
)

type PaymentController struct {
	app.Controller
	PaymentService *services.PaymentService
}

func NewPaymentController() *PaymentController {
	paymentService := services.NewPaymentService()
	return &PaymentController{
		PaymentService: paymentService,
	}
}

func (self *PaymentController) Deposit(w http.ResponseWriter, r *http.Request) {
	var body requests.DepositRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		self.JsonValidationErrors(w, err)
		return
	}

	validate := validator.New()
	err = validate.Struct(body)
	if err != nil {
		self.JsonValidationErrors(w, err)
		return
	}

	params := types.DepositParams{
		Amount:           body.Amount,
		Currency:         body.Currency,
		Token:            body.Token,
		TransactionId:    body.TransactionId,
		UserId:           body.UserId,
		Provider:         body.Provider,
		CreditCardNumber: body.CreditCardNumber,
		ExpirationDate:   body.ExpirationDate,
		CVV:              body.CVV,
	}
	if body.Provider == "stripe" {
		self.PaymentService.SetPaymentProvider(providers.NewStripePaymentProvider())
	} else if body.Provider == "authorize" {
		self.PaymentService.SetPaymentProvider(providers.NewAuthorizeNetPaymentProvider())

	} else {
		self.JsonError(w, "Invalid provider", http.StatusBadRequest)
		return
	}
	res, err := self.PaymentService.Deposit(params)
	if err != nil {
		statusCode := errors.MapErrorToStatusCode(err)
		self.JsonError(w, err.Error(), statusCode)
		return
	}
	self.Json(w, res, http.StatusOK)
}

func (self *PaymentController) Withdraw(w http.ResponseWriter, r *http.Request) {
	var body requests.WithdrawRequest
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		self.JsonValidationErrors(w, err)
		return
	}

	validate := validator.New()
	err = validate.Struct(body)
	if err != nil {
		self.JsonValidationErrors(w, err)
		return
	}

	params := types.WithdrawParams{
		Amount:           body.Amount,
		Currency:         body.Currency,
		Destination:      body.Destination,
		TransactionId:    body.TransactionId,
		UserId:           body.UserId,
		Provider:         body.Provider,
		CreditCardNumber: body.CreditCardNumber,
		ExpirationDate:   body.ExpirationDate,
		CVV:              body.CVV,
	}

	if body.Provider == "stripe" {
		self.PaymentService.SetPaymentProvider(providers.NewStripePaymentProvider())
	} else if body.Provider == "authorize" {
		self.PaymentService.SetPaymentProvider(providers.NewAuthorizeNetPaymentProvider())
	} else {
		self.JsonError(w, "Invalid provider", http.StatusBadRequest)
		return
	}
	res, err := self.PaymentService.Withdraw(params)
	if err != nil {
		statusCode := errors.MapErrorToStatusCode(err)
		self.JsonError(w, err.Error(), statusCode)
		return
	}
	self.Json(w, res, http.StatusOK)
}

func (self *PaymentController) StripeWebhook(w http.ResponseWriter, r *http.Request) {

	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Error reading request body: %v", err)
		return
	}

	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), app.App().Config().GetString("payment.stripe_endpoint_secret"))
	if err != nil {
		app.App().Logger().Error("Error verifying webhook signature: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = self.PaymentService.HandleStripeEvents(event)
	if err != nil {
		app.App().Logger().Error("Error handling stripe event: ", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	self.Json(w, nil, http.StatusOK)
}

func (self *PaymentController) AuthorizeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !self.PaymentService.VerifyAuthorizeSignature(r, body) {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
	}

	app.App().Logger().Info("Received Webhook: ", string(body))

	var event requests.WebhookEvent
	err = json.Unmarshal(body, &event)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	self.PaymentService.HandleAuthorizeEvents(event)

	// Respond to the webhook request
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))

	self.Json(w, nil, http.StatusOK)
}
