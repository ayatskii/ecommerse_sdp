package factory

import (
	"fmt"

	"github.com/ecommerce/payment-system/internal/strategy"
	"github.com/ecommerce/payment-system/pkg/errors"
)

type StrategyFactory struct {
	supportedStrategies map[string]bool
}

func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{
		supportedStrategies: map[string]bool{
			"instant":  true,
			"deferred": true,
			"split":    true,
		},
	}
}

func (f *StrategyFactory) CreateStrategy(strategyType string, params map[string]interface{}) (strategy.PaymentStrategy, error) {
	if !f.supportedStrategies[strategyType] {
		return nil, errors.NewValidationError(
			fmt.Sprintf("unsupported payment strategy: %s", strategyType),
		)
	}

	switch strategyType {
	case "instant":
		return f.createInstantStrategy(params)
	case "deferred":
		return f.createDeferredStrategy(params)
	case "split":
		return nil, errors.NewValidationError("split strategy must be created with CreateSplitStrategy")
	default:
		return nil, errors.NewValidationError(
			fmt.Sprintf("unsupported payment strategy: %s", strategyType),
		)
	}
}

func (f *StrategyFactory) createInstantStrategy(params map[string]interface{}) (strategy.PaymentStrategy, error) {
	minAmount := 1.0
	maxAmount := 10000.0

	if val, ok := params["min_amount"].(float64); ok {
		minAmount = val
	}
	if val, ok := params["max_amount"].(float64); ok {
		maxAmount = val
	}

	return strategy.NewInstantPaymentStrategy(minAmount, maxAmount), nil
}

func (f *StrategyFactory) createDeferredStrategy(params map[string]interface{}) (strategy.PaymentStrategy, error) {
	minAmount := 100.0
	maxAmount := 10000.0
	installments := 3
	interestRate := 0.0

	if val, ok := params["min_amount"].(float64); ok {
		minAmount = val
	}
	if val, ok := params["max_amount"].(float64); ok {
		maxAmount = val
	}
	if val, ok := params["installments"].(int); ok {
		installments = val
	} else if val, ok := params["installments"].(float64); ok {
		installments = int(val)
	}
	if val, ok := params["interest_rate"].(float64); ok {
		interestRate = val
	}

	if installments < 2 {
		return nil, errors.NewValidationError("deferred payment requires at least 2 installments")
	}
	if installments > 12 {
		return nil, errors.NewValidationError("deferred payment cannot exceed 12 installments")
	}

	return strategy.NewDeferredPaymentStrategy(minAmount, maxAmount, installments, interestRate), nil
}

func (f *StrategyFactory) CreateSplitStrategy(payments []strategy.SplitPaymentItem) (strategy.PaymentStrategy, error) {
	return strategy.NewSplitPaymentStrategy(payments)
}

func (f *StrategyFactory) IsSupported(strategyType string) bool {
	return f.supportedStrategies[strategyType]
}

func (f *StrategyFactory) GetSupportedStrategies() []string {
	strategies := make([]string, 0, len(f.supportedStrategies))
	for s := range f.supportedStrategies {
		strategies = append(strategies, s)
	}
	return strategies
}
