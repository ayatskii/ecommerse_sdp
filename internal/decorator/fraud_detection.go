package decorator

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type FraudDetectionDecorator struct {
	*BaseDecorator
	maxRiskScore             int
	velocityCheckWindow      time.Duration
	maxTransactionsPerWindow int
	transactionHistory       map[string][]time.Time
	mu                       sync.RWMutex
}

type FraudDetectionConfig struct {
	MaxRiskScore             int
	VelocityCheckWindow      time.Duration
	MaxTransactionsPerWindow int
	CustomerID               string
}

func NewFraudDetectionDecorator(wrapped payment.Payment, config FraudDetectionConfig) *FraudDetectionDecorator {
	return &FraudDetectionDecorator{
		BaseDecorator:            NewBaseDecorator(wrapped),
		maxRiskScore:             config.MaxRiskScore,
		velocityCheckWindow:      config.VelocityCheckWindow,
		maxTransactionsPerWindow: config.MaxTransactionsPerWindow,
		transactionHistory:       make(map[string][]time.Time),
	}
}

func (d *FraudDetectionDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Applying fraud detection decorator",
		zap.Float64("amount", amount),
	)

	riskScore := d.calculateRiskScore(amount)

	logger.Info("Fraud risk calculated",
		zap.Int("risk_score", riskScore),
		zap.Int("max_risk_score", d.maxRiskScore),
	)

	if riskScore > d.maxRiskScore {
		return nil, errors.NewFraudDetectedError(
			fmt.Sprintf("transaction blocked: high fraud risk (score: %d)", riskScore),
		)
	}

	if err := d.velocityCheck(); err != nil {
		return nil, err
	}

	if err := d.geolocationCheck(); err != nil {
		return nil, err
	}

	result, err := d.wrapped.Process(ctx, amount)
	if err != nil {
		return nil, err
	}

	d.recordTransaction()

	result.AppliedDecorators = append(result.AppliedDecorators, "fraud_detection")

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["fraud_risk_score"] = riskScore
	result.Metadata["fraud_checks_passed"] = []string{
		"risk_score",
		"velocity_check",
		"geolocation_check",
	}

	return result, nil
}

func (d *FraudDetectionDecorator) calculateRiskScore(amount float64) int {

	score := 0

	if amount > 1000 {
		score += 20
	}
	if amount > 5000 {
		score += 30
	}

	score += rand.Intn(30)

	return score
}

func (d *FraudDetectionDecorator) velocityCheck() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	customerID := "default"

	transactions, exists := d.transactionHistory[customerID]
	if !exists {
		return nil
	}

	cutoff := time.Now().Add(-d.velocityCheckWindow)
	recent := []time.Time{}
	for _, tx := range transactions {
		if tx.After(cutoff) {
			recent = append(recent, tx)
		}
	}

	if len(recent) >= d.maxTransactionsPerWindow {
		return errors.NewFraudDetectedError(
			fmt.Sprintf("transaction velocity exceeded: %d transactions in %v",
				len(recent), d.velocityCheckWindow),
		)
	}

	return nil
}

func (d *FraudDetectionDecorator) geolocationCheck() error {

	if rand.Intn(100) < 5 {
		return errors.NewFraudDetectedError("geolocation validation failed")
	}

	return nil
}

func (d *FraudDetectionDecorator) recordTransaction() {
	d.mu.Lock()
	defer d.mu.Unlock()

	customerID := "default"

	transactions, exists := d.transactionHistory[customerID]
	if !exists {
		transactions = []time.Time{}
	}

	transactions = append(transactions, time.Now())
	d.transactionHistory[customerID] = transactions
}
