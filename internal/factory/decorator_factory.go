package factory

import (
	"fmt"
	"time"

	"github.com/ecommerce/payment-system/config"
	"github.com/ecommerce/payment-system/internal/decorator"
	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type DecoratorFactory struct {
	config *config.Config
}

func NewDecoratorFactory(cfg *config.Config) *DecoratorFactory {
	return &DecoratorFactory{
		config: cfg,
	}
}

func (f *DecoratorFactory) CreateDecoratorChain(
	basePayment payment.Payment,
	features []string,
	options domain.CheckoutOptions,
	customer *domain.Customer,
) (payment.Payment, error) {
	logger.Info("Building decorator chain",
		zap.Strings("features", features),
		zap.String("payment_type", basePayment.GetType()),
	)

	current := basePayment

	for _, feature := range features {
		var err error
		current, err = f.createDecorator(feature, current, options, customer)
		if err != nil {
			return nil, fmt.Errorf("failed to create decorator %s: %w", feature, err)
		}
		logger.Debug("Decorator applied",
			zap.String("decorator", feature),
		)
	}

	return current, nil
}

func (f *DecoratorFactory) createDecorator(
	feature string,
	wrapped payment.Payment,
	options domain.CheckoutOptions,
	customer *domain.Customer,
) (payment.Payment, error) {
	switch feature {
	case "discount":
		return f.createDiscountDecorator(wrapped, options)
	case "cashback":
		return f.createCashbackDecorator(wrapped)
	case "fraud_detection":
		return f.createFraudDetectionDecorator(wrapped, customer)
	case "tax":
		return f.createTaxDecorator(wrapped, customer)
	case "loyalty_points":
		return f.createLoyaltyPointsDecorator(wrapped, options, customer)
	default:
		return nil, errors.NewValidationError(fmt.Sprintf("unsupported decorator: %s", feature))
	}
}

func (f *DecoratorFactory) createDiscountDecorator(
	wrapped payment.Payment,
	options domain.CheckoutOptions,
) (payment.Payment, error) {
	if !f.config.Decorators.Discount.Enabled {
		return wrapped, nil
	}

	config := decorator.DiscountConfig{
		DiscountType:  "percentage",
		DiscountValue: 10.0,
		MinAmount:     0,
		MaxDiscount:   f.config.Decorators.Discount.MaxFixedAmount,
		ExpiryDate:    time.Now().Add(30 * 24 * time.Hour),
		DiscountCode:  options.DiscountCode,
	}

	return decorator.NewDiscountDecorator(wrapped, config)
}

func (f *DecoratorFactory) createCashbackDecorator(wrapped payment.Payment) (payment.Payment, error) {
	if !f.config.Decorators.Cashback.Enabled {
		return wrapped, nil
	}

	config := decorator.CashbackConfig{
		Tier1Threshold:  f.config.Decorators.Cashback.Tier1Threshold,
		Tier1Percentage: f.config.Decorators.Cashback.Tier1Percentage,
		Tier2Percentage: f.config.Decorators.Cashback.Tier2Percentage,
	}

	return decorator.NewCashbackDecorator(wrapped, config), nil
}

func (f *DecoratorFactory) createFraudDetectionDecorator(
	wrapped payment.Payment,
	customer *domain.Customer,
) (payment.Payment, error) {
	if !f.config.Decorators.FraudDetection.Enabled {
		return wrapped, nil
	}

	customerID := ""
	if customer != nil {
		customerID = customer.ID
	}

	config := decorator.FraudDetectionConfig{
		MaxRiskScore:             f.config.Decorators.FraudDetection.MaxRiskScore,
		VelocityCheckWindow:      f.config.Decorators.FraudDetection.VelocityCheckWindow,
		MaxTransactionsPerWindow: f.config.Decorators.FraudDetection.MaxTransactionsPerWindow,
		CustomerID:               customerID,
	}

	return decorator.NewFraudDetectionDecorator(wrapped, config), nil
}

func (f *DecoratorFactory) createTaxDecorator(
	wrapped payment.Payment,
	customer *domain.Customer,
) (payment.Payment, error) {
	if !f.config.Decorators.Tax.Enabled {
		return wrapped, nil
	}

	region := "DEFAULT"
	if customer != nil && customer.Address.State != "" {
		region = customer.Address.State
	}

	config := decorator.TaxConfig{
		Region:      region,
		TaxRates:    f.config.Decorators.Tax.Rates,
		DefaultRate: f.config.Decorators.Tax.DefaultRate,
	}

	return decorator.NewTaxDecorator(wrapped, config), nil
}

func (f *DecoratorFactory) createLoyaltyPointsDecorator(
	wrapped payment.Payment,
	options domain.CheckoutOptions,
	customer *domain.Customer,
) (payment.Payment, error) {
	if !f.config.Decorators.LoyaltyPoints.Enabled {
		return wrapped, nil
	}

	if customer == nil || options.UseLoyaltyPoints == 0 {
		return wrapped, nil
	}

	config := decorator.LoyaltyPointsConfig{
		AvailablePoints:         customer.LoyaltyPoints,
		PointsToRedeem:          options.UseLoyaltyPoints,
		PointsToCurrencyRatio:   f.config.Decorators.LoyaltyPoints.PointsToCurrencyRatio,
		MaxRedemptionPercentage: f.config.Decorators.LoyaltyPoints.MaxRedemptionPercentage,
	}

	return decorator.NewLoyaltyPointsDecorator(wrapped, config)
}

func (f *DecoratorFactory) GetAvailableDecorators() []string {
	decorators := []string{}

	if f.config.Decorators.Discount.Enabled {
		decorators = append(decorators, "discount")
	}
	if f.config.Decorators.Cashback.Enabled {
		decorators = append(decorators, "cashback")
	}
	if f.config.Decorators.FraudDetection.Enabled {
		decorators = append(decorators, "fraud_detection")
	}
	if f.config.Decorators.Tax.Enabled {
		decorators = append(decorators, "tax")
	}
	if f.config.Decorators.LoyaltyPoints.Enabled {
		decorators = append(decorators, "loyalty_points")
	}

	return decorators
}
