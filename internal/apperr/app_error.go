package apperr

import (
	"fmt"
	"net/http"
)

type AppError struct {
	Msg    string
	Err    error
	Status int
}

func NewAppError(msg string, code int, err error) error {
	return &AppError{Msg: msg, Err: err, Status: code}
}

func NewInternalServerError(msg string, err error) error {
	return &AppError{Msg: msg, Err: err, Status: http.StatusInternalServerError}
}

func NewBadRequestError(msg string, err error) error {
	return &AppError{Msg: msg, Err: err, Status: http.StatusBadRequest}
}

func NewAuthError(msg string, err error) error {
	return &AppError{Msg: msg, Err: err, Status: http.StatusUnauthorized}
}

func (e AppError) Error() string {
	return fmt.Sprintf("%v", e.Msg)
}
