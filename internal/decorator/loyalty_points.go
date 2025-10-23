package decorator

import (
	"context"
	"fmt"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type LoyaltyPointsDecorator struct {
	*BaseDecorator
	availablePoints         int
	pointsToRedeem          int
	pointsToCurrencyRatio   float64
	maxRedemptionPercentage float64
}

type LoyaltyPointsConfig struct {
	AvailablePoints         int
	PointsToRedeem          int
	PointsToCurrencyRatio   float64
	MaxRedemptionPercentage float64
}

func NewLoyaltyPointsDecorator(wrapped payment.Payment, config LoyaltyPointsConfig) (*LoyaltyPointsDecorator, error) {
	if config.PointsToRedeem > config.AvailablePoints {
		return nil, errors.NewValidationError("insufficient loyalty points")
	}

	if config.PointsToRedeem < 0 {
		return nil, errors.NewValidationError("points to redeem cannot be negative")
	}

	return &LoyaltyPointsDecorator{
		BaseDecorator:           NewBaseDecorator(wrapped),
		availablePoints:         config.AvailablePoints,
		pointsToRedeem:          config.PointsToRedeem,
		pointsToCurrencyRatio:   config.PointsToCurrencyRatio,
		maxRedemptionPercentage: config.MaxRedemptionPercentage,
	}, nil
}

func (d *LoyaltyPointsDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Applying loyalty points decorator",
		zap.Float64("amount", amount),
		zap.Int("points_to_redeem", d.pointsToRedeem),
		zap.Int("available_points", d.availablePoints),
	)

	discount := float64(d.pointsToRedeem) / d.pointsToCurrencyRatio

	maxRedemption := amount * (d.maxRedemptionPercentage / 100.0)
	if discount > maxRedemption {
		return nil, errors.NewValidationError(
			fmt.Sprintf("loyalty points redemption exceeds maximum (%.2f%% of purchase)",
				d.maxRedemptionPercentage),
		)
	}

	finalAmount := amount - discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	pointsEarned := int(amount)

	logger.Info("Loyalty points processed",
		zap.Float64("original_amount", amount),
		zap.Float64("discount", discount),
		zap.Float64("final_amount", finalAmount),
		zap.Int("points_redeemed", d.pointsToRedeem),
		zap.Int("points_earned", pointsEarned),
	)

	result, err := d.wrapped.Process(ctx, finalAmount)
	if err != nil {
		return nil, err
	}

	if result.OriginalAmount == 0 {
		result.OriginalAmount = amount
	}
	result.ProcessedAmount = finalAmount
	result.Amount = finalAmount
	result.AppliedDecorators = append(result.AppliedDecorators, "loyalty_points")

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["loyalty_points_redeemed"] = d.pointsToRedeem
	result.Metadata["loyalty_points_earned"] = pointsEarned
	result.Metadata["loyalty_discount"] = discount
	result.Metadata["loyalty_balance_after"] = d.availablePoints - d.pointsToRedeem + pointsEarned

	return result, nil
}
