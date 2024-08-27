package app

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type Controller struct {
}

func Json(res http.ResponseWriter, payload interface{}, statusCode int) {
	response, _ := json.Marshal(payload)
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	_, _ = res.Write(response)
}

func JsonError(res http.ResponseWriter, msg string, statusCode int, name ...string) {
	response := map[string]interface{}{
		"success": false,
		"message": msg,
	}

	Json(res, response, statusCode)
}

type CustomHttpError struct {
	StatusCode int
	Message    string
	Name       string
	Error      *error
}

func (c *Controller) JsonError(res http.ResponseWriter, msg string, statusCode int) {
	JsonError(res, msg, statusCode)
}

func (c *Controller) Json(res http.ResponseWriter, payload interface{}, statusCode int) {
	Json(res, payload, statusCode)
}

type CustomFieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
	Tag   string `json:"tag"`
}

func (c *Controller) JsonValidationErrors(res http.ResponseWriter, err error) {
	var validationErrors validator.ValidationErrors
	errors.As(err, &validationErrors)

	var fieldErrors = make([]CustomFieldError, 0)
	if len(validationErrors) > 0 {
		for _, fieldErr := range validationErrors {
			fieldErrors = append(fieldErrors, CustomFieldError{
				Field: fieldErr.Field(),
				Error: fieldErr.Error(),
				Tag:   fieldErr.Tag(),
			})
		}
	}

	Json(res, map[string]interface{}{
		"status":      http.StatusBadRequest,
		"type":        "https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html",
		"title":       "Validation Error",
		"message":     err.Error(),
		"errors":      err,
		"fieldErrors": fieldErrors,
	}, http.StatusBadRequest)

}
