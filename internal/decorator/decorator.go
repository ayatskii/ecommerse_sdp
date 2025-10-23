package decorator

import (
	"context"

	"github.com/ecommerce/payment-system/internal/payment"
)

type PaymentDecorator interface {
	payment.Payment
	GetWrapped() payment.Payment
}

type BaseDecorator struct {
	wrapped payment.Payment
}

func NewBaseDecorator(wrapped payment.Payment) *BaseDecorator {
	return &BaseDecorator{wrapped: wrapped}
}

func (d *BaseDecorator) GetWrapped() payment.Payment {
	return d.wrapped
}

func (d *BaseDecorator) GetType() string {
	return d.wrapped.GetType()
}

func (d *BaseDecorator) GetDetails() map[string]interface{} {
	return d.wrapped.GetDetails()
}

func (d *BaseDecorator) Process(ctx context.Context, amount float64) (*payment.PaymentResult, error) {
	return d.wrapped.Process(ctx, amount)
}
