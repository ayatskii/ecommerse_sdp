package decorator

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type DiscountDecorator struct {
	*BaseDecorator
	discountType  string
	discountValue float64
	minAmount     float64
	maxDiscount   float64
	expiryDate    time.Time
	discountCode  string
}

type DiscountConfig struct {
	DiscountType  string
	DiscountValue float64
	MinAmount     float64
	MaxDiscount   float64
	ExpiryDate    time.Time
	DiscountCode  string
}

func NewDiscountDecorator(wrapped payment.Payment, config DiscountConfig) (*DiscountDecorator, error) {
	if config.DiscountValue <= 0 {
		return nil, errors.NewValidationError("discount value must be positive")
	}

	if config.DiscountType == "percentage" && config.DiscountValue > 100 {
		return nil, errors.NewValidationError("percentage discount cannot exceed 100%")
	}

	return &DiscountDecorator{
		BaseDecorator: NewBaseDecorator(wrapped),
		discountType:  config.DiscountType,
		discountValue: config.DiscountValue,
		minAmount:     config.MinAmount,
		maxDiscount:   config.MaxDiscount,
		expiryDate:    config.ExpiryDate,
		discountCode:  config.DiscountCode,
	}, nil
}

func (d *DiscountDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Applying discount decorator",
		zap.String("type", d.discountType),
		zap.Float64("value", d.discountValue),
		zap.Float64("original_amount", amount),
	)

	if !d.expiryDate.IsZero() && time.Now().After(d.expiryDate) {
		return nil, errors.NewValidationError("discount code has expired")
	}

	if amount < d.minAmount {
		return nil, errors.NewValidationError(
			fmt.Sprintf("minimum amount for discount is $%.2f", d.minAmount),
		)
	}

	discountAmount := d.calculateDiscount(amount)
	finalAmount := amount - discountAmount

	if finalAmount < 0 {
		finalAmount = 0
	}

	logger.Info("Discount applied",
		zap.Float64("original_amount", amount),
		zap.Float64("discount_amount", discountAmount),
		zap.Float64("final_amount", finalAmount),
	)

	result, err := d.wrapped.Process(ctx, finalAmount)
	if err != nil {
		return nil, err
	}

	result.OriginalAmount = amount
	result.ProcessedAmount = finalAmount
	result.Amount = finalAmount
	result.AppliedDecorators = append(result.AppliedDecorators, "discount")

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["discount_type"] = d.discountType
	result.Metadata["discount_value"] = d.discountValue
	result.Metadata["discount_amount"] = discountAmount
	result.Metadata["discount_code"] = d.discountCode

	return result, nil
}

func (d *DiscountDecorator) calculateDiscount(amount float64) float64 {
	var discount float64

	if d.discountType == "percentage" {
		discount = amount * (d.discountValue / 100.0)
	} else {
		discount = d.discountValue
	}

	if d.maxDiscount > 0 && discount > d.maxDiscount {
		discount = d.maxDiscount
	}

	if discount > amount {
		discount = amount
	}

	return discount
}
