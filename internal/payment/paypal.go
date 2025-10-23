package payment

import (
	"context"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"github.com/ecommerce/payment-system/pkg/validator"
	"go.uber.org/zap"
)

type PayPalPayment struct {
	email     string
	password  string
	validator *validator.EmailValidator
}

func NewPayPalPayment(email, password string) (*PayPalPayment, error) {
	v := validator.NewEmailValidator()

	if err := v.Validate(email); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidPayment, "invalid PayPal email")
	}

	if password == "" {
		return nil, errors.NewInvalidPaymentError("PayPal password is required")
	}

	return &PayPalPayment{
		email:     email,
		password:  password,
		validator: v,
	}, nil
}

func (p *PayPalPayment) Process(ctx context.Context, amount float64) (*PaymentResult, error) {
	logger.Info("Processing PayPal payment",
		zap.Float64("amount", amount),
		zap.String("email", p.email),
	)

	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), errors.ErrCodeTimeout, "payment context expired")
	}

	amountValidator := validator.NewAmountValidator()
	if err := amountValidator.Validate(amount, 1.0, 5000.0); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeValidation, "invalid payment amount")
	}

	time.Sleep(150 * time.Millisecond)

	transactionID := domain.NewID()

	result := &PaymentResult{
		Success:         true,
		TransactionID:   transactionID,
		Amount:          amount,
		OriginalAmount:  amount,
		ProcessedAmount: amount,
		Currency:        "USD",
		PaymentMethod:   "paypal",
		Message:         "PayPal payment processed successfully",
		Metadata: map[string]interface{}{
			"paypal_email": p.email,
			"processed_at": time.Now().Format(time.RFC3339),
		},
		AppliedDecorators: []string{},
	}

	logger.Info("PayPal payment processed successfully",
		zap.String("transaction_id", transactionID),
		zap.Float64("amount", amount),
	)

	return result, nil
}

func (p *PayPalPayment) GetType() string {
	return "paypal"
}

func (p *PayPalPayment) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"type":         "paypal",
		"paypal_email": p.email,
	}
}
