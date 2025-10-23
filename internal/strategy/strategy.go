package strategy

import (
	"context"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/payment"
)

type PaymentStrategy interface {
	Execute(ctx context.Context, payment payment.Payment, amount float64) (*payment.PaymentResult, error)
	GetName() string
	ValidateAmount(amount float64) error
}

type PaymentContext struct {
	strategy PaymentStrategy
}

func NewPaymentContext(strategy PaymentStrategy) *PaymentContext {
	return &PaymentContext{
		strategy: strategy,
	}
}

func (c *PaymentContext) SetStrategy(strategy PaymentStrategy) {
	c.strategy = strategy
}

func (c *PaymentContext) ExecutePayment(ctx context.Context, payment payment.Payment, amount float64) (*payment.PaymentResult, error) {
	return c.strategy.Execute(ctx, payment, amount)
}

type SplitPaymentItem struct {
	Payment payment.Payment
	Amount  float64
}

type DeferredPaymentSchedule struct {
	ID           string                       `json:"id"`
	TotalAmount  float64                      `json:"total_amount"`
	Installments int                          `json:"installments"`
	InterestRate float64                      `json:"interest_rate"`
	Payments     []DeferredPaymentInstallment `json:"payments"`
}

type DeferredPaymentInstallment struct {
	InstallmentNumber int     `json:"installment_number"`
	Amount            float64 `json:"amount"`
	DueDate           string  `json:"due_date"`
	Status            string  `json:"status"`
}

func CreateDeferredSchedule(amount float64, installments int, interestRate float64) *DeferredPaymentSchedule {
	schedule := &DeferredPaymentSchedule{
		ID:           domain.NewID(),
		TotalAmount:  amount,
		Installments: installments,
		InterestRate: interestRate,
		Payments:     make([]DeferredPaymentInstallment, 0, installments),
	}

	totalWithInterest := amount * (1 + interestRate/100)
	installmentAmount := totalWithInterest / float64(installments)

	for i := 0; i < installments; i++ {
		schedule.Payments = append(schedule.Payments, DeferredPaymentInstallment{
			InstallmentNumber: i + 1,
			Amount:            installmentAmount,
			DueDate:           "",
			Status:            "pending",
		})
	}

	return schedule
}
