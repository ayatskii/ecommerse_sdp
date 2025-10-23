package service

import (
	"context"
	"fmt"

	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type InventoryService struct {
	repo repository.Repository
}

func NewInventoryService(repo repository.Repository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) CheckAvailability(ctx context.Context, productID string, quantity int) (bool, error) {
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return false, err
	}

	available := product.Stock >= quantity

	logger.Debug("Inventory check",
		zap.String("product_id", productID),
		zap.Int("requested", quantity),
		zap.Int("available", product.Stock),
		zap.Bool("sufficient", available),
	)

	return available, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, productID string, quantity int) error {
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors.NewInventoryError(
			fmt.Sprintf("insufficient stock for product %s: have %d, need %d",
				product.Name, product.Stock, quantity),
		)
	}

	product.Stock -= quantity

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	logger.Info("Stock reserved",
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
		zap.Int("remaining", product.Stock),
	)

	return nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, productID string, quantity int) error {
	product, err := s.repo.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	product.Stock += quantity

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	logger.Info("Stock released",
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
		zap.Int("new_stock", product.Stock),
	)

	return nil
}
