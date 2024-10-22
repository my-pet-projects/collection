package model

import "fmt"

type AppError struct {
	Msg string
	Err error
}

func NewAppError(msg string, err error) error {
	return &AppError{Msg: msg, Err: err}
}

func (e AppError) Error() string {
	return fmt.Sprintf("%v", e.Msg)
}
