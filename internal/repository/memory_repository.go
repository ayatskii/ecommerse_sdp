package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/errors"
)

type MemoryRepository struct {
	customers    map[string]*domain.Customer
	products     map[string]*domain.Product
	carts        map[string]*domain.Cart
	transactions map[string]*domain.Transaction
	mu           sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		customers:    make(map[string]*domain.Customer),
		products:     make(map[string]*domain.Product),
		carts:        make(map[string]*domain.Cart),
		transactions: make(map[string]*domain.Transaction),
	}

	repo.seedData()

	return repo
}

func (r *MemoryRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customers[customer.ID]; exists {
		return errors.NewAlreadyExistsError("customer")
	}

	r.customers[customer.ID] = customer
	return nil
}

func (r *MemoryRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	customer, exists := r.customers[id]
	if !exists {
		return nil, errors.NewNotFoundError("customer")
	}

	return customer, nil
}

func (r *MemoryRepository) GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, customer := range r.customers {
		if customer.Email == email {
			return customer, nil
		}
	}

	return nil, errors.NewNotFoundError("customer")
}

func (r *MemoryRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.customers[customer.ID]; !exists {
		return errors.NewNotFoundError("customer")
	}

	r.customers[customer.ID] = customer
	return nil
}

func (r *MemoryRepository) ListCustomers(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	customers := make([]*domain.Customer, 0, len(r.customers))
	for _, c := range r.customers {
		customers = append(customers, c)
	}

	start := offset
	end := offset + limit

	if start >= len(customers) {
		return []*domain.Customer{}, nil
	}
	if end > len(customers) {
		end = len(customers)
	}

	return customers[start:end], nil
}

func (r *MemoryRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; exists {
		return errors.NewAlreadyExistsError("product")
	}

	r.products[product.ID] = product
	return nil
}

func (r *MemoryRepository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, errors.NewNotFoundError("product")
	}

	return product, nil
}

func (r *MemoryRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; !exists {
		return errors.NewNotFoundError("product")
	}

	r.products[product.ID] = product
	return nil
}

func (r *MemoryRepository) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]*domain.Product, 0, len(r.products))
	for _, p := range r.products {
		products = append(products, p)
	}

	start := offset
	end := offset + limit

	if start >= len(products) {
		return []*domain.Product{}, nil
	}
	if end > len(products) {
		end = len(products)
	}

	return products[start:end], nil
}

func (r *MemoryRepository) CreateCart(ctx context.Context, cart *domain.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[cart.ID]; exists {
		return errors.NewAlreadyExistsError("cart")
	}

	r.carts[cart.ID] = cart
	return nil
}

func (r *MemoryRepository) GetCart(ctx context.Context, id string) (*domain.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cart, exists := r.carts[id]
	if !exists {
		return nil, errors.NewNotFoundError("cart")
	}

	return cart, nil
}

func (r *MemoryRepository) UpdateCart(ctx context.Context, cart *domain.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.carts[cart.ID]; !exists {
		return errors.NewNotFoundError("cart")
	}

	r.carts[cart.ID] = cart
	return nil
}

func (r *MemoryRepository) GetCartByCustomer(ctx context.Context, customerID string) (*domain.Cart, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, cart := range r.carts {
		if cart.CustomerID == customerID {
			return cart, nil
		}
	}

	return nil, errors.NewNotFoundError("cart")
}

func (r *MemoryRepository) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.transactions[transaction.ID]; exists {
		return errors.NewAlreadyExistsError("transaction")
	}

	r.transactions[transaction.ID] = transaction
	return nil
}

func (r *MemoryRepository) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transaction, exists := r.transactions[id]
	if !exists {
		return nil, errors.NewNotFoundError("transaction")
	}

	return transaction, nil
}

func (r *MemoryRepository) ListTransactionsByCustomer(ctx context.Context, customerID string, limit, offset int) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transactions := make([]*domain.Transaction, 0)
	for _, t := range r.transactions {
		if t.CustomerID == customerID {
			transactions = append(transactions, t)
		}
	}

	start := offset
	end := offset + limit

	if start >= len(transactions) {
		return []*domain.Transaction{}, nil
	}
	if end > len(transactions) {
		end = len(transactions)
	}

	return transactions[start:end], nil
}

func (r *MemoryRepository) Close() error {

	return nil
}

func (r *MemoryRepository) seedData() {

	products := []*domain.Product{
		{
			ID:          "prod-1",
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999.99,
			SKU:         "LAP-001",
			Stock:       10,
			Category:    "Electronics",
		},
		{
			ID:          "prod-2",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			SKU:         "MOU-001",
			Stock:       50,
			Category:    "Accessories",
		},
		{
			ID:          "prod-3",
			Name:        "USB-C Cable",
			Description: "High-speed USB-C cable",
			Price:       19.99,
			SKU:         "CAB-001",
			Stock:       100,
			Category:    "Accessories",
		},
		{
			ID:          "prod-4",
			Name:        "Mechanical Keyboard",
			Description: "RGB mechanical keyboard",
			Price:       149.99,
			SKU:         "KEY-001",
			Stock:       25,
			Category:    "Accessories",
		},
		{
			ID:          "prod-5",
			Name:        "Monitor",
			Description: "27-inch 4K monitor",
			Price:       399.99,
			SKU:         "MON-001",
			Stock:       15,
			Category:    "Electronics",
		},
	}

	for _, p := range products {
		r.products[p.ID] = p
	}

	customer := &domain.Customer{
		ID:            "cust-1",
		Email:         "john.doe@example.com",
		Name:          "John Doe",
		Phone:         "+1234567890",
		LoyaltyPoints: 500,
		Address: domain.Address{
			Street:     "123 Main St",
			City:       "San Francisco",
			State:      "CA",
			PostalCode: "94105",
			Country:    "USA",
		},
	}

	r.customers[customer.ID] = customer

	fmt.Println("✓ Sample data seeded successfully")
}
