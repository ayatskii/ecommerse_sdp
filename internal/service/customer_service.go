package service

import (
	"context"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type CustomerService struct {
	repo repository.Repository
}

func NewCustomerService(repo repository.Repository) *CustomerService {
	return &CustomerService{repo: repo}
}

func (s *CustomerService) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	return s.repo.GetCustomer(ctx, id)
}

func (s *CustomerService) UpdateLoyaltyPoints(ctx context.Context, customerID string, earned, redeemed int) error {
	customer, err := s.repo.GetCustomer(ctx, customerID)
	if err != nil {
		return err
	}

	customer.LoyaltyPoints = customer.LoyaltyPoints + earned - redeemed

	if err := s.repo.UpdateCustomer(ctx, customer); err != nil {
		return err
	}

	logger.Info("Loyalty points updated",
		zap.String("customer_id", customerID),
		zap.Int("earned", earned),
		zap.Int("redeemed", redeemed),
		zap.Int("new_balance", customer.LoyaltyPoints),
	)

	return nil
}
