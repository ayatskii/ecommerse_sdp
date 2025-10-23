package service

import (
	"context"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type TransactionService struct {
	repo repository.Repository
}

func NewTransactionService(repo repository.Repository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	if err := s.repo.CreateTransaction(ctx, transaction); err != nil {
		return err
	}

	logger.Info("Transaction created",
		zap.String("transaction_id", transaction.ID),
		zap.String("customer_id", transaction.CustomerID),
		zap.Float64("amount", transaction.Amount),
		zap.String("status", string(transaction.Status)),
	)

	return nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	return s.repo.GetTransaction(ctx, id)
}

func (s *TransactionService) GetCustomerTransactions(ctx context.Context, customerID string, limit, offset int) ([]*domain.Transaction, error) {
	return s.repo.ListTransactionsByCustomer(ctx, customerID, limit, offset)
}
