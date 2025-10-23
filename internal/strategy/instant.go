package strategy

import (
	"context"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"github.com/ecommerce/payment-system/pkg/validator"
	"go.uber.org/zap"
)

type InstantPaymentStrategy struct {
	minAmount float64
	maxAmount float64
}

func NewInstantPaymentStrategy(minAmount, maxAmount float64) *InstantPaymentStrategy {
	return &InstantPaymentStrategy{
		minAmount: minAmount,
		maxAmount: maxAmount,
	}
}

func (s *InstantPaymentStrategy) Execute(ctx context.Context, payment payment.Payment, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Executing instant payment strategy",
		zap.String("payment_type", payment.GetType()),
		zap.Float64("amount", amount),
	)

	if err := s.ValidateAmount(amount); err != nil {
		return nil, err
	}

	result, err := payment.Process(ctx, amount)
	if err != nil {
		logger.Error("Instant payment failed",
			zap.Error(err),
			zap.Float64("amount", amount),
		)
		return nil, errors.Wrap(err, errors.ErrCodePaymentFailed, "instant payment processing failed")
	}

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["payment_strategy"] = "instant"

	logger.Info("Instant payment completed successfully",
		zap.String("transaction_id", result.TransactionID),
		zap.Float64("amount", amount),
	)

	return result, nil
}

func (s *InstantPaymentStrategy) GetName() string {
	return "instant"
}

func (s *InstantPaymentStrategy) ValidateAmount(amount float64) error {
	v := validator.NewAmountValidator()
	return v.Validate(amount, s.minAmount, s.maxAmount)
}
