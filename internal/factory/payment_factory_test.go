package factory

import (
	"testing"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaymentFactory(t *testing.T) {
	factory := NewPaymentFactory()

	t.Run("Create Credit Card Payment", func(t *testing.T) {
		config := payment.PaymentConfig{
			CardNumber: "4532015112830366",
			CardHolder: "John Doe",
			ExpiryDate: "12/25",
			CVV:        "123",
		}

		p, err := factory.CreatePayment("credit_card", config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "credit_card", p.GetType())
	})

	t.Run("Create PayPal Payment", func(t *testing.T) {
		config := payment.PaymentConfig{
			PayPalEmail:    "user@example.com",
			PayPalPassword: "password",
		}

		p, err := factory.CreatePayment("paypal", config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "paypal", p.GetType())
	})

	t.Run("Create Crypto Payment", func(t *testing.T) {
		config := payment.PaymentConfig{
			WalletAddress: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
			CryptoType:    "BTC",
		}

		p, err := factory.CreatePayment("crypto", config)
		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "crypto", p.GetType())
	})

	t.Run("Unsupported Payment Type", func(t *testing.T) {
		config := payment.PaymentConfig{}
		_, err := factory.CreatePayment("unsupported", config)
		assert.Error(t, err)
	})

	t.Run("Get Supported Types", func(t *testing.T) {
		types := factory.GetSupportedTypes()
		assert.Contains(t, types, "credit_card")
		assert.Contains(t, types, "paypal")
		assert.Contains(t, types, "crypto")
	})
}
