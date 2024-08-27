package routes

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"payment-service/app"
	"payment-service/controllers"
)

func GetRoutes() []app.Route {
	var appRoutes []app.Route
	appRoutes = append(appRoutes, PaymentRoutes...)
	return appRoutes
}

var paymentController = *controllers.NewPaymentController()
var PaymentRoutes = []app.Route{
	{"Post", "/api/v1/deposit", nil, paymentController.Deposit},
	{"Post", "/api/v1/withdraw", nil, paymentController.Withdraw},
	{"Post", "/api/v1/stripe-webhook", nil, paymentController.StripeWebhook},
	{"Post", "/api/v1/authorize-webhook", nil, paymentController.StripeWebhook},
	{"GET", "/swagger.json", nil, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./swagger.json")
	}},
	{"GET", "/swagger/*", nil, httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger.json"),
	)},
}
