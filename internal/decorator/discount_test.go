package decorator

import (
	"context"
	"testing"
	"time"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscountDecorator(t *testing.T) {
	basePayment, _ := payment.NewCreditCardPayment(
		"4532015112830366",
		"John Doe",
		"12/25",
		"123",
	)

	t.Run("Percentage Discount", func(t *testing.T) {
		config := DiscountConfig{
			DiscountType:  "percentage",
			DiscountValue: 10.0,
			MinAmount:     0,
			MaxDiscount:   100,
			ExpiryDate:    time.Now().Add(24 * time.Hour),
			DiscountCode:  "TEST10",
		}

		decorator, err := NewDiscountDecorator(basePayment, config)
		require.NoError(t, err)

		ctx := context.Background()
		result, err := decorator.Process(ctx, 100.00)
		require.NoError(t, err)

		assert.True(t, result.Success)
		assert.Equal(t, 100.00, result.OriginalAmount)
		assert.Equal(t, 90.00, result.ProcessedAmount)
		assert.Contains(t, result.AppliedDecorators, "discount")
	})

	t.Run("Fixed Discount", func(t *testing.T) {
		config := DiscountConfig{
			DiscountType:  "fixed",
			DiscountValue: 20.0,
			MinAmount:     0,
			MaxDiscount:   100,
			ExpiryDate:    time.Now().Add(24 * time.Hour),
			DiscountCode:  "SAVE20",
		}

		decorator, err := NewDiscountDecorator(basePayment, config)
		require.NoError(t, err)

		ctx := context.Background()
		result, err := decorator.Process(ctx, 100.00)
		require.NoError(t, err)

		assert.Equal(t, 80.00, result.ProcessedAmount)
	})

	t.Run("Expired Discount", func(t *testing.T) {
		config := DiscountConfig{
			DiscountType:  "percentage",
			DiscountValue: 10.0,
			ExpiryDate:    time.Now().Add(-24 * time.Hour),
		}

		decorator, err := NewDiscountDecorator(basePayment, config)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = decorator.Process(ctx, 100.00)
		assert.Error(t, err)
	})
}
