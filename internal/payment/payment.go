package payment

import (
	"context"
)

type Payment interface {
	Process(ctx context.Context, amount float64) (*PaymentResult, error)
	GetType() string
	GetDetails() map[string]interface{}
}

type PaymentResult struct {
	Success           bool                   `json:"success"`
	TransactionID     string                 `json:"transaction_id"`
	Amount            float64                `json:"amount"`
	OriginalAmount    float64                `json:"original_amount"`
	ProcessedAmount   float64                `json:"processed_amount"`
	Currency          string                 `json:"currency"`
	PaymentMethod     string                 `json:"payment_method"`
	Message           string                 `json:"message"`
	Metadata          map[string]interface{} `json:"metadata"`
	AppliedDecorators []string               `json:"applied_decorators"`
}

type PaymentConfig struct {
	Currency string
	Metadata map[string]interface{}

	CardNumber string
	CardHolder string
	ExpiryDate string
	CVV        string

	PayPalEmail    string
	PayPalPassword string

	WalletAddress string
	CryptoType    string
}
