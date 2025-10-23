package payment

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreditCardPayment(t *testing.T) {
	t.Run("Valid Credit Card", func(t *testing.T) {
		payment, err := NewCreditCardPayment(
			"4532015112830366",
			"John Doe",
			"12/25",
			"123",
		)
		require.NoError(t, err)
		assert.NotNil(t, payment)
	})

	t.Run("Invalid Card Number", func(t *testing.T) {
		_, err := NewCreditCardPayment(
			"1234567890",
			"John Doe",
			"12/25",
			"123",
		)
		assert.Error(t, err)
	})

	t.Run("Invalid CVV", func(t *testing.T) {
		_, err := NewCreditCardPayment(
			"4532015112830366",
			"John Doe",
			"12/25",
			"12",
		)
		assert.Error(t, err)
	})

	t.Run("Process Payment", func(t *testing.T) {
		payment, err := NewCreditCardPayment(
			"4532015112830366",
			"John Doe",
			"12/25",
			"123",
		)
		require.NoError(t, err)

		ctx := context.Background()
		result, err := payment.Process(ctx, 100.00)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 100.00, result.Amount)
		assert.Equal(t, "credit_card", result.PaymentMethod)
	})
}
