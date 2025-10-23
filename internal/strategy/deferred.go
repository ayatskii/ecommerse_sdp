package strategy

import (
	"context"
	"fmt"

	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"github.com/ecommerce/payment-system/pkg/validator"
	"go.uber.org/zap"
)

type DeferredPaymentStrategy struct {
	minAmount    float64
	maxAmount    float64
	installments int
	interestRate float64
}

func NewDeferredPaymentStrategy(minAmount, maxAmount float64, installments int, interestRate float64) *DeferredPaymentStrategy {
	return &DeferredPaymentStrategy{
		minAmount:    minAmount,
		maxAmount:    maxAmount,
		installments: installments,
		interestRate: interestRate,
	}
}

func (s *DeferredPaymentStrategy) Execute(ctx context.Context, payment payment.Payment, amount float64) (*payment.PaymentResult, error) {
	logger.Info("Executing deferred payment strategy",
		zap.String("payment_type", payment.GetType()),
		zap.Float64("amount", amount),
		zap.Int("installments", s.installments),
	)

	if err := s.ValidateAmount(amount); err != nil {
		return nil, err
	}

	schedule := CreateDeferredSchedule(amount, s.installments, s.interestRate)

	firstInstallment := schedule.Payments[0].Amount

	logger.Info("Processing first installment",
		zap.Float64("first_installment", firstInstallment),
		zap.Float64("total_amount", amount),
		zap.Int("installments", s.installments),
	)

	result, err := payment.Process(ctx, firstInstallment)
	if err != nil {
		logger.Error("Deferred payment first installment failed",
			zap.Error(err),
			zap.Float64("amount", firstInstallment),
		)
		return nil, errors.Wrap(err, errors.ErrCodePaymentFailed, "deferred payment processing failed")
	}

	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["payment_strategy"] = "deferred"
	result.Metadata["schedule_id"] = schedule.ID
	result.Metadata["total_amount"] = schedule.TotalAmount
	result.Metadata["installments"] = s.installments
	result.Metadata["interest_rate"] = s.interestRate
	result.Metadata["first_installment"] = firstInstallment
	result.Metadata["remaining_installments"] = s.installments - 1

	result.OriginalAmount = amount
	result.Amount = firstInstallment
	result.ProcessedAmount = firstInstallment

	logger.Info("Deferred payment first installment completed",
		zap.String("transaction_id", result.TransactionID),
		zap.String("schedule_id", schedule.ID),
		zap.Int("remaining_installments", s.installments-1),
	)

	return result, nil
}

func (s *DeferredPaymentStrategy) GetName() string {
	return fmt.Sprintf("deferred_%d_installments", s.installments)
}

func (s *DeferredPaymentStrategy) ValidateAmount(amount float64) error {
	v := validator.NewAmountValidator()
	return v.Validate(amount, s.minAmount, s.maxAmount)
}
