package errors

import (
	"errors"
	"fmt"
)

const (
	ErrCodeValidation        = "VALIDATION_ERROR"
	ErrCodeNotFound          = "NOT_FOUND"
	ErrCodeAlreadyExists     = "ALREADY_EXISTS"
	ErrCodeUnauthorized      = "UNAUTHORIZED"
	ErrCodeInternalError     = "INTERNAL_ERROR"
	ErrCodePaymentFailed     = "PAYMENT_FAILED"
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeInvalidPayment    = "INVALID_PAYMENT"
	ErrCodeFraudDetected     = "FRAUD_DETECTED"
	ErrCodeInventoryError    = "INVENTORY_ERROR"
	ErrCodeTimeout           = "TIMEOUT"
)

type AppError struct {
	Code    string
	Message string
	Err     error
	Details map[string]interface{}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

func Wrap(err error, code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Details: make(map[string]interface{}),
	}
}

func (e *AppError) WithDetails(key string, value interface{}) *AppError {
	e.Details[key] = value
	return e
}

func NewValidationError(message string) *AppError {
	return New(ErrCodeValidation, message)
}

func NewNotFoundError(resource string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s not found", resource))
}

func NewAlreadyExistsError(resource string) *AppError {
	return New(ErrCodeAlreadyExists, fmt.Sprintf("%s already exists", resource))
}

func NewUnauthorizedError(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

func NewInternalError(message string) *AppError {
	return New(ErrCodeInternalError, message)
}

func NewPaymentError(message string) *AppError {
	return New(ErrCodePaymentFailed, message)
}

func NewInsufficientFundsError() *AppError {
	return New(ErrCodeInsufficientFunds, "insufficient funds")
}

func NewInvalidPaymentError(message string) *AppError {
	return New(ErrCodeInvalidPayment, message)
}

func NewFraudDetectedError(message string) *AppError {
	return New(ErrCodeFraudDetected, message)
}

func NewInventoryError(message string) *AppError {
	return New(ErrCodeInventoryError, message)
}

func NewTimeoutError(message string) *AppError {
	return New(ErrCodeTimeout, message)
}

func IsErrorCode(err error, code string) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func GetErrorCode(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeInternalError
}
