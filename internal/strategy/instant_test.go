package strategy

import (
	"context"
	"testing"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstantPaymentStrategy(t *testing.T) {
	strategy := NewInstantPaymentStrategy(1.0, 10000.0)

	basePayment, _ := payment.NewCreditCardPayment(
		"4532015112830366",
		"John Doe",
		"12/25",
		"123",
	)

	t.Run("Successful Payment", func(t *testing.T) {
		ctx := context.Background()
		result, err := strategy.Execute(ctx, basePayment, 100.00)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 100.00, result.Amount)
	})

	t.Run("Amount Below Minimum", func(t *testing.T) {
		ctx := context.Background()
		_, err := strategy.Execute(ctx, basePayment, 0.50)
		assert.Error(t, err)
	})

	t.Run("Amount Above Maximum", func(t *testing.T) {
		ctx := context.Background()
		_, err := strategy.Execute(ctx, basePayment, 15000.00)
		assert.Error(t, err)
	})
}
