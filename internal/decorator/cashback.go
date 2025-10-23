package decorator

import (
	"context"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type CashbackDecorator struct {
	*BaseDecorator
	tier1Threshold  float64
	tier1Percentage float64
	tier2Percentage float64
}

type CashbackConfig struct {
	Tier1Threshold  float64
	Tier1Percentage float64
	Tier2Percentage float64
}

func NewCashbackDecorator(wrapped payment.Payment, config CashbackConfig) *CashbackDecorator {
	return &CashbackDecorator{
		BaseDecorator:   NewBaseDecorator(wrapped),
		tier1Threshold:  config.Tier1Threshold,
		tier1Percentage: config.Tier1Percentage,
		tier2Percentage: config.Tier2Percentage,
	}
}

func (d *CashbackDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Applying cashback decorator",
		zap.Float64("amount", amount),
	)

	cashbackAmount := d.calculateCashback(amount)

	logger.Info("Cashback calculated",
		zap.Float64("amount", amount),
		zap.Float64("cashback_amount", cashbackAmount),
	)

	result, err := d.wrapped.Process(ctx, amount)
	if err != nil {
		return nil, err
	}

	result.AppliedDecorators = append(result.AppliedDecorators, "cashback")

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["cashback_amount"] = cashbackAmount
	result.Metadata["cashback_percentage"] = d.getCashbackPercentage(amount)

	return result, nil
}

func (d *CashbackDecorator) calculateCashback(amount float64) float64 {
	percentage := d.getCashbackPercentage(amount)
	return amount * (percentage / 100.0)
}

func (d *CashbackDecorator) getCashbackPercentage(amount float64) float64 {
	if amount >= d.tier1Threshold {
		return d.tier2Percentage
	}
	return d.tier1Percentage
}
