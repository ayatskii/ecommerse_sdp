package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"github.com/ecommerce/payment-system/pkg/validator"
	"go.uber.org/zap"
)

type CreditCardPayment struct {
	cardNumber string
	cardHolder string
	expiryDate string
	cvv        string
	validator  *validator.CreditCardValidator
}

func NewCreditCardPayment(cardNumber, cardHolder, expiryDate, cvv string) (*CreditCardPayment, error) {
	v := validator.NewCreditCardValidator()

	if err := v.ValidateCardNumber(cardNumber); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidPayment, "invalid card number")
	}

	if err := v.ValidateCVV(cvv); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidPayment, "invalid CVV")
	}

	if err := v.ValidateExpiryDate(expiryDate); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidPayment, "invalid expiry date")
	}

	if cardHolder == "" {
		return nil, errors.NewInvalidPaymentError("card holder name is required")
	}

	return &CreditCardPayment{
		cardNumber: cardNumber,
		cardHolder: cardHolder,
		expiryDate: expiryDate,
		cvv:        cvv,
		validator:  v,
	}, nil
}

func (p *CreditCardPayment) Process(ctx context.Context, amount float64) (*PaymentResult, error) {
	logger.Info("Processing credit card payment",
		zap.Float64("amount", amount),
		zap.String("card_holder", p.cardHolder),
	)

	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), errors.ErrCodeTimeout, "payment context expired")
	}

	amountValidator := validator.NewAmountValidator()
	if err := amountValidator.Validate(amount, 1.0, 10000.0); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeValidation, "invalid payment amount")
	}

	time.Sleep(100 * time.Millisecond)

	transactionID := domain.NewID()

	result := &PaymentResult{
		Success:         true,
		TransactionID:   transactionID,
		Amount:          amount,
		OriginalAmount:  amount,
		ProcessedAmount: amount,
		Currency:        "USD",
		PaymentMethod:   "credit_card",
		Message:         "Payment processed successfully",
		Metadata: map[string]interface{}{
			"card_holder":   p.cardHolder,
			"last_4_digits": p.getLastFourDigits(),
			"processed_at":  time.Now().Format(time.RFC3339),
		},
		AppliedDecorators: []string{},
	}

	logger.Info("Credit card payment processed successfully",
		zap.String("transaction_id", transactionID),
		zap.Float64("amount", amount),
	)

	return result, nil
}

func (p *CreditCardPayment) GetType() string {
	return "credit_card"
}

func (p *CreditCardPayment) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"type":          "credit_card",
		"card_holder":   p.cardHolder,
		"last_4_digits": p.getLastFourDigits(),
		"expiry_date":   p.expiryDate,
	}
}

func (p *CreditCardPayment) getLastFourDigits() string {
	if len(p.cardNumber) < 4 {
		return "****"
	}
	return fmt.Sprintf("****%s", p.cardNumber[len(p.cardNumber)-4:])
}
