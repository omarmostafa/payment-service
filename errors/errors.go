package errors

import (
	"net/http"
)

type ValidationError struct {
	Message string `json:"message"`
}

type InternalServerError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func (e *InternalServerError) Error() string {
	return e.Message
}

func MapErrorToStatusCode(err error) int {
	switch err.(type) {
	case *ValidationError:
		return http.StatusBadRequest
	case *InternalServerError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
