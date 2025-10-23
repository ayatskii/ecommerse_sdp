package strategy

import (
	"context"
	"fmt"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type SplitPaymentStrategy struct {
	payments []SplitPaymentItem
}

func NewSplitPaymentStrategy(payments []SplitPaymentItem) (*SplitPaymentStrategy, error) {
	if len(payments) == 0 {
		return nil, errors.NewValidationError("at least one payment method is required")
	}

	if len(payments) > 5 {
		return nil, errors.NewValidationError("maximum 5 payment methods allowed for split payment")
	}

	return &SplitPaymentStrategy{
		payments: payments,
	}, nil
}

func (s *SplitPaymentStrategy) Execute(ctx context.Context, _ payment.Payment, totalAmount float64) (*payment.PaymentResult, error) {
	logger.Info("Executing split payment strategy",
		zap.Float64("total_amount", totalAmount),
		zap.Int("payment_methods", len(s.payments)),
	)

	if err := s.ValidateAmount(totalAmount); err != nil {
		return nil, err
	}

	var splitSum float64
	for _, item := range s.payments {
		splitSum += item.Amount
	}

	if fmt.Sprintf("%.2f", splitSum) != fmt.Sprintf("%.2f", totalAmount) {
		return nil, errors.NewValidationError(
			fmt.Sprintf("split payment amounts (%.2f) do not match total amount (%.2f)",
				splitSum, totalAmount),
		)
	}

	var processedResults []*payment.PaymentResult
	var totalProcessed float64

	for i, item := range s.payments {
		logger.Info("Processing split payment part",
			zap.Int("part", i+1),
			zap.Int("total_parts", len(s.payments)),
			zap.Float64("amount", item.Amount),
			zap.String("payment_type", item.Payment.GetType()),
		)

		result, err := item.Payment.Process(ctx, item.Amount)
		if err != nil {

			s.rollbackPayments(ctx, processedResults)
			return nil, errors.Wrap(err, errors.ErrCodePaymentFailed,
				fmt.Sprintf("split payment part %d failed", i+1))
		}

		processedResults = append(processedResults, result)
		totalProcessed += item.Amount
	}

	combinedResult := &payment.PaymentResult{
		Success:         true,
		TransactionID:   processedResults[0].TransactionID,
		Amount:          totalAmount,
		OriginalAmount:  totalAmount,
		ProcessedAmount: totalProcessed,
		Currency:        "USD",
		PaymentMethod:   "split",
		Message:         fmt.Sprintf("Split payment completed across %d methods", len(s.payments)),
		Metadata: map[string]interface{}{
			"payment_strategy": "split",
			"payment_count":    len(s.payments),
			"split_details":    s.getSplitDetails(processedResults),
		},
		AppliedDecorators: []string{},
	}

	logger.Info("Split payment completed successfully",
		zap.String("transaction_id", combinedResult.TransactionID),
		zap.Float64("total_amount", totalAmount),
		zap.Int("payment_methods", len(s.payments)),
	)

	return combinedResult, nil
}

func (s *SplitPaymentStrategy) GetName() string {
	return fmt.Sprintf("split_%d_methods", len(s.payments))
}

func (s *SplitPaymentStrategy) ValidateAmount(amount float64) error {
	if amount <= 0 {
		return errors.NewValidationError("amount must be positive")
	}

	for i, item := range s.payments {
		if item.Amount <= 0 {
			return errors.NewValidationError(
				fmt.Sprintf("split payment part %d has invalid amount: %.2f", i+1, item.Amount),
			)
		}
	}

	return nil
}

func (s *SplitPaymentStrategy) rollbackPayments(ctx context.Context, processedResults []*payment.PaymentResult) {
	logger.Warn("Rolling back split payments",
		zap.Int("count", len(processedResults)),
	)

	for i, result := range processedResults {
		logger.Info("Rolling back payment",
			zap.Int("part", i+1),
			zap.String("transaction_id", result.TransactionID),
			zap.Float64("amount", result.Amount),
		)

	}
}

func (s *SplitPaymentStrategy) getSplitDetails(results []*payment.PaymentResult) []map[string]interface{} {
	details := make([]map[string]interface{}, 0, len(results))

	for i, result := range results {
		details = append(details, map[string]interface{}{
			"part":           i + 1,
			"payment_method": result.PaymentMethod,
			"amount":         result.Amount,
			"transaction_id": result.TransactionID,
			"status":         "completed",
		})
	}

	return details
}
