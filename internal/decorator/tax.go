package decorator

import (
	"context"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type TaxDecorator struct {
	*BaseDecorator
	region      string
	taxRate     float64
	taxRates    map[string]float64
	defaultRate float64
}

type TaxConfig struct {
	Region      string
	TaxRates    map[string]float64
	DefaultRate float64
}

func NewTaxDecorator(wrapped payment.Payment, config TaxConfig) *TaxDecorator {
	decorator := &TaxDecorator{
		BaseDecorator: NewBaseDecorator(wrapped),
		region:        config.Region,
		taxRates:      config.TaxRates,
		defaultRate:   config.DefaultRate,
	}

	if rate, exists := config.TaxRates[config.Region]; exists {
		decorator.taxRate = rate
	} else {
		decorator.taxRate = config.DefaultRate
	}

	return decorator
}

func (d *TaxDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Applying tax decorator",
		zap.Float64("amount", amount),
		zap.String("region", d.region),
		zap.Float64("tax_rate", d.taxRate),
	)

	taxAmount := amount * (d.taxRate / 100.0)
	totalAmount := amount + taxAmount

	logger.Info("Tax calculated",
		zap.Float64("subtotal", amount),
		zap.Float64("tax_amount", taxAmount),
		zap.Float64("total_amount", totalAmount),
	)

	result, err := d.wrapped.Process(ctx, totalAmount)
	if err != nil {
		return nil, err
	}

	if result.OriginalAmount == 0 {
		result.OriginalAmount = amount
	}
	result.ProcessedAmount = totalAmount
	result.Amount = totalAmount
	result.AppliedDecorators = append(result.AppliedDecorators, "tax")

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["subtotal"] = amount
	result.Metadata["tax_amount"] = taxAmount
	result.Metadata["tax_rate"] = d.taxRate
	result.Metadata["tax_region"] = d.region

	return result, nil
}
