package service

import (
	"context"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type CartService struct {
	repo repository.Repository
}

func NewCartService(repo repository.Repository) *CartService {
	return &CartService{repo: repo}
}

func (s *CartService) CreateCart(ctx context.Context, customerID string) (*domain.Cart, error) {
	cart := &domain.Cart{
		ID:         domain.NewID(),
		CustomerID: customerID,
		Items:      []domain.CartItem{},
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.CreateCart(ctx, cart); err != nil {
		return nil, err
	}

	logger.Info("Cart created",
		zap.String("cart_id", cart.ID),
		zap.String("customer_id", customerID),
	)

	return cart, nil
}

func (s *CartService) GetOrCreateCart(ctx context.Context, customerID string) (*domain.Cart, error) {
	cart, err := s.repo.GetCartByCustomer(ctx, customerID)
	if err == nil {
		return cart, nil
	}

	return s.CreateCart(ctx, customerID)
}

func (s *CartService) AddItem(ctx context.Context, cartID string, product *domain.Product, quantity int) error {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		return err
	}

	cart.AddItem(*product, quantity)
	cart.UpdatedAt = time.Now()

	if err := s.repo.UpdateCart(ctx, cart); err != nil {
		return err
	}

	logger.Info("Item added to cart",
		zap.String("cart_id", cartID),
		zap.String("product_id", product.ID),
		zap.Int("quantity", quantity),
	)

	return nil
}

func (s *CartService) RemoveItem(ctx context.Context, cartID, productID string) error {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		return err
	}

	cart.RemoveItem(productID)
	cart.UpdatedAt = time.Now()

	if err := s.repo.UpdateCart(ctx, cart); err != nil {
		return err
	}

	logger.Info("Item removed from cart",
		zap.String("cart_id", cartID),
		zap.String("product_id", productID),
	)

	return nil
}

func (s *CartService) UpdateQuantity(ctx context.Context, cartID, productID string, quantity int) error {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		return err
	}

	cart.UpdateQuantity(productID, quantity)
	cart.UpdatedAt = time.Now()

	if err := s.repo.UpdateCart(ctx, cart); err != nil {
		return err
	}

	logger.Info("Cart item quantity updated",
		zap.String("cart_id", cartID),
		zap.String("product_id", productID),
		zap.Int("quantity", quantity),
	)

	return nil
}

func (s *CartService) ClearCart(ctx context.Context, cartID string) error {
	cart, err := s.repo.GetCart(ctx, cartID)
	if err != nil {
		return err
	}

	cart.Clear()
	cart.UpdatedAt = time.Now()

	if err := s.repo.UpdateCart(ctx, cart); err != nil {
		return err
	}

	logger.Info("Cart cleared",
		zap.String("cart_id", cartID),
	)

	return nil
}
