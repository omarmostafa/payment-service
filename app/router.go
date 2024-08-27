package app

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Route struct {
	Method      string
	Pattern     string
	Middlewares *chi.Middlewares
	HandlerFunc func(res http.ResponseWriter, r *http.Request)
}
type Routes []Route
