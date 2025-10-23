package payment

import (
	"context"
	"strings"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"github.com/ecommerce/payment-system/pkg/validator"
	"go.uber.org/zap"
)

type CryptoPayment struct {
	walletAddress string
	cryptoType    string
	validator     *validator.CryptoAddressValidator
}

func NewCryptoPayment(walletAddress, cryptoType string) (*CryptoPayment, error) {
	v := validator.NewCryptoAddressValidator()

	cryptoType = strings.ToUpper(cryptoType)

	if err := v.Validate(walletAddress, cryptoType); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInvalidPayment, "invalid wallet address")
	}

	supportedTypes := map[string]bool{
		"BTC":  true,
		"ETH":  true,
		"USDT": true,
	}

	if !supportedTypes[cryptoType] {
		return nil, errors.NewInvalidPaymentError("unsupported cryptocurrency type: " + cryptoType)
	}

	return &CryptoPayment{
		walletAddress: walletAddress,
		cryptoType:    cryptoType,
		validator:     v,
	}, nil
}

func (p *CryptoPayment) Process(ctx context.Context, amount float64) (*PaymentResult, error) {
	logger.Info("Processing crypto payment",
		zap.Float64("amount", amount),
		zap.String("crypto_type", p.cryptoType),
		zap.String("wallet_address", p.maskWalletAddress()),
	)

	if ctx.Err() != nil {
		return nil, errors.Wrap(ctx.Err(), errors.ErrCodeTimeout, "payment context expired")
	}

	amountValidator := validator.NewAmountValidator()
	if err := amountValidator.Validate(amount, 10.0, 50000.0); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeValidation, "invalid payment amount")
	}

	time.Sleep(200 * time.Millisecond)

	transactionID := domain.NewID()

	result := &PaymentResult{
		Success:         true,
		TransactionID:   transactionID,
		Amount:          amount,
		OriginalAmount:  amount,
		ProcessedAmount: amount,
		Currency:        p.cryptoType,
		PaymentMethod:   "crypto",
		Message:         "Cryptocurrency payment processed successfully",
		Metadata: map[string]interface{}{
			"crypto_type":    p.cryptoType,
			"wallet_address": p.maskWalletAddress(),
			"blockchain_tx":  "0x" + transactionID[:16],
			"processed_at":   time.Now().Format(time.RFC3339),
		},
		AppliedDecorators: []string{},
	}

	logger.Info("Crypto payment processed successfully",
		zap.String("transaction_id", transactionID),
		zap.Float64("amount", amount),
		zap.String("crypto_type", p.cryptoType),
	)

	return result, nil
}

func (p *CryptoPayment) GetType() string {
	return "crypto"
}

func (p *CryptoPayment) GetDetails() map[string]interface{} {
	return map[string]interface{}{
		"type":           "crypto",
		"crypto_type":    p.cryptoType,
		"wallet_address": p.maskWalletAddress(),
	}
}

func (p *CryptoPayment) maskWalletAddress() string {
	if len(p.walletAddress) < 10 {
		return "****"
	}
	return p.walletAddress[:6] + "****" + p.walletAddress[len(p.walletAddress)-4:]
}
