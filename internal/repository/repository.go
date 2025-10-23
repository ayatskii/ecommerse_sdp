package repository

import (
	"context"

	"github.com/ecommerce/payment-system/internal/domain"
)

type Repository interface {
	CreateCustomer(ctx context.Context, customer *domain.Customer) error
	GetCustomer(ctx context.Context, id string) (*domain.Customer, error)
	GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, customer *domain.Customer) error
	ListCustomers(ctx context.Context, limit, offset int) ([]*domain.Customer, error)

	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error)

	CreateCart(ctx context.Context, cart *domain.Cart) error
	GetCart(ctx context.Context, id string) (*domain.Cart, error)
	UpdateCart(ctx context.Context, cart *domain.Cart) error
	GetCartByCustomer(ctx context.Context, customerID string) (*domain.Cart, error)

	CreateTransaction(ctx context.Context, transaction *domain.Transaction) error
	GetTransaction(ctx context.Context, id string) (*domain.Transaction, error)
	ListTransactionsByCustomer(ctx context.Context, customerID string, limit, offset int) ([]*domain.Transaction, error)

	Close() error
}
