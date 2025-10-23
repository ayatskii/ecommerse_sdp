package factory

import (
	"fmt"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
)

type PaymentFactory struct {
	supportedTypes map[string]bool
}

func NewPaymentFactory() *PaymentFactory {
	return &PaymentFactory{
		supportedTypes: map[string]bool{
			"credit_card": true,
			"paypal":      true,
			"crypto":      true,
		},
	}
}

func (f *PaymentFactory) CreatePayment(paymentType string, config payment.PaymentConfig) (payment.Payment, error) {

	if !f.supportedTypes[paymentType] {
		return nil, errors.NewInvalidPaymentError(
			fmt.Sprintf("unsupported payment type: %s", paymentType),
		)
	}

	switch paymentType {
	case "credit_card":
		return f.createCreditCardPayment(config)
	case "paypal":
		return f.createPayPalPayment(config)
	case "crypto":
		return f.createCryptoPayment(config)
	default:
		return nil, errors.NewInvalidPaymentError(
			fmt.Sprintf("unsupported payment type: %s", paymentType),
		)
	}
}

func (f *PaymentFactory) createCreditCardPayment(config payment.PaymentConfig) (payment.Payment, error) {

	if config.CardNumber == "" {
		return nil, errors.NewValidationError("card number is required")
	}
	if config.CardHolder == "" {
		return nil, errors.NewValidationError("card holder is required")
	}
	if config.ExpiryDate == "" {
		return nil, errors.NewValidationError("expiry date is required")
	}
	if config.CVV == "" {
		return nil, errors.NewValidationError("CVV is required")
	}

	return payment.NewCreditCardPayment(
		config.CardNumber,
		config.CardHolder,
		config.ExpiryDate,
		config.CVV,
	)
}

func (f *PaymentFactory) createPayPalPayment(config payment.PaymentConfig) (payment.Payment, error) {

	if config.PayPalEmail == "" {
		return nil, errors.NewValidationError("PayPal email is required")
	}
	if config.PayPalPassword == "" {
		return nil, errors.NewValidationError("PayPal password is required")
	}

	return payment.NewPayPalPayment(
		config.PayPalEmail,
		config.PayPalPassword,
	)
}

func (f *PaymentFactory) createCryptoPayment(config payment.PaymentConfig) (payment.Payment, error) {

	if config.WalletAddress == "" {
		return nil, errors.NewValidationError("wallet address is required")
	}
	if config.CryptoType == "" {
		return nil, errors.NewValidationError("crypto type is required")
	}

	return payment.NewCryptoPayment(
		config.WalletAddress,
		config.CryptoType,
	)
}

func (f *PaymentFactory) IsSupported(paymentType string) bool {
	return f.supportedTypes[paymentType]
}

func (f *PaymentFactory) GetSupportedTypes() []string {
	types := make([]string, 0, len(f.supportedTypes))
	for t := range f.supportedTypes {
		types = append(types, t)
	}
	return types
}
