package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ecommerce/payment-system/internal/domain"
)

type FileRepository struct {
	*MemoryRepository
	filePath string
	mu       sync.RWMutex
}

type PersistentData struct {
	Customers    map[string]*domain.Customer    `json:"customers"`
	Products     map[string]*domain.Product     `json:"products"`
	Carts        map[string]*domain.Cart        `json:"carts"`
	Transactions map[string]*domain.Transaction `json:"transactions"`
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	memRepo := NewMemoryRepository()

	repo := &FileRepository{
		MemoryRepository: memRepo,
		filePath:         filePath,
	}

	if err := repo.load(); err != nil {
		fmt.Printf("⚠ Could not load data from file, using fresh data: %v\n", err)
	} else {
		fmt.Println("✓ Data loaded from file")
	}

	return repo, nil
}

func (r *FileRepository) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var persistentData PersistentData
	if err := json.Unmarshal(data, &persistentData); err != nil {
		return err
	}

	if len(persistentData.Customers) > 0 {
		r.customers = persistentData.Customers
	}
	if len(persistentData.Products) > 0 {
		r.products = persistentData.Products
	}
	if len(persistentData.Carts) > 0 {
		r.carts = persistentData.Carts
	}
	if len(persistentData.Transactions) > 0 {
		r.transactions = persistentData.Transactions
	}

	return nil
}

func (r *FileRepository) save() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	persistentData := PersistentData{
		Customers:    r.customers,
		Products:     r.products,
		Carts:        r.carts,
		Transactions: r.transactions,
	}

	data, err := json.MarshalIndent(persistentData, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	return os.WriteFile(r.filePath, data, 0644)
}

func (r *FileRepository) CreateCart(ctx context.Context, cart *domain.Cart) error {
	if err := r.MemoryRepository.CreateCart(ctx, cart); err != nil {
		return err
	}
	return r.save()
}

func (r *FileRepository) UpdateCart(ctx context.Context, cart *domain.Cart) error {
	if err := r.MemoryRepository.UpdateCart(ctx, cart); err != nil {
		return err
	}
	return r.save()
}

func (r *FileRepository) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	if err := r.MemoryRepository.CreateTransaction(ctx, transaction); err != nil {
		return err
	}
	return r.save()
}

func (r *FileRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	if err := r.MemoryRepository.UpdateCustomer(ctx, customer); err != nil {
		return err
	}
	return r.save()
}

func (r *FileRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	if err := r.MemoryRepository.UpdateProduct(ctx, product); err != nil {
		return err
	}
	return r.save()
}

func (r *FileRepository) Close() error {
	return r.save()
}
