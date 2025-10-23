package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/errors"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &SQLiteRepository{db: db}

	if err := repo.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	if err := repo.seedData(); err != nil {
		return nil, fmt.Errorf("failed to seed data: %w", err)
	}

	return repo, nil
}

func (r *SQLiteRepository) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS customers (
		id TEXT PRIMARY KEY,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		phone TEXT,
		loyalty_points INTEGER DEFAULT 0,
		address_street TEXT,
		address_city TEXT,
		address_state TEXT,
		address_postal_code TEXT,
		address_country TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS products (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		sku TEXT UNIQUE NOT NULL,
		stock INTEGER DEFAULT 0,
		category TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS carts (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		items TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (customer_id) REFERENCES customers(id)
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id TEXT PRIMARY KEY,
		customer_id TEXT NOT NULL,
		amount REAL NOT NULL,
		status TEXT NOT NULL,
		payment_method TEXT NOT NULL,
		payment_details TEXT,
		metadata TEXT,
		error_message TEXT,
		processed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (customer_id) REFERENCES customers(id)
	);

	CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
	CREATE INDEX IF NOT EXISTS idx_carts_customer ON carts(customer_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_customer ON transactions(customer_id);
	`

	_, err := r.db.Exec(schema)
	return err
}

func (r *SQLiteRepository) seedData() error {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	products := []*domain.Product{
		{
			ID:          "prod-1",
			Name:        "Laptop",
			Description: "High-performance laptop",
			Price:       999.99,
			SKU:         "LAP-001",
			Stock:       10,
			Category:    "Electronics",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "prod-2",
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			SKU:         "MOU-001",
			Stock:       50,
			Category:    "Accessories",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "prod-3",
			Name:        "USB-C Cable",
			Description: "High-speed USB-C cable",
			Price:       19.99,
			SKU:         "CAB-001",
			Stock:       100,
			Category:    "Accessories",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "prod-4",
			Name:        "Mechanical Keyboard",
			Description: "RGB mechanical keyboard",
			Price:       149.99,
			SKU:         "KEY-001",
			Stock:       25,
			Category:    "Accessories",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "prod-5",
			Name:        "Monitor",
			Description: "27-inch 4K monitor",
			Price:       399.99,
			SKU:         "MON-001",
			Stock:       15,
			Category:    "Electronics",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, p := range products {
		if err := r.CreateProduct(context.Background(), p); err != nil {
			return err
		}
	}

	defaultCustomer := &domain.Customer{
		ID:            "cust-default",
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
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.CreateCustomer(context.Background(), defaultCustomer); err != nil {
		return err
	}

	fmt.Println("✓ Sample data seeded successfully")
	fmt.Printf("✓ Default user created: %s\n", defaultCustomer.Email)
	return nil
}

func (r *SQLiteRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customers (id, email, name, phone, loyalty_points, 
			address_street, address_city, address_state, address_postal_code, address_country,
			created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		customer.ID, customer.Email, customer.Name, customer.Phone, customer.LoyaltyPoints,
		customer.Address.Street, customer.Address.City, customer.Address.State,
		customer.Address.PostalCode, customer.Address.Country,
		customer.CreatedAt, customer.UpdatedAt,
	)

	return err
}

func (r *SQLiteRepository) GetCustomer(ctx context.Context, id string) (*domain.Customer, error) {
	query := `
		SELECT id, email, name, phone, loyalty_points,
			address_street, address_city, address_state, address_postal_code, address_country,
			created_at, updated_at
		FROM customers WHERE id = ?
	`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&customer.ID, &customer.Email, &customer.Name, &customer.Phone, &customer.LoyaltyPoints,
		&customer.Address.Street, &customer.Address.City, &customer.Address.State,
		&customer.Address.PostalCode, &customer.Address.Country,
		&customer.CreatedAt, &customer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("customer")
	}

	return customer, err
}

func (r *SQLiteRepository) GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT id, email, name, phone, loyalty_points,
			address_street, address_city, address_state, address_postal_code, address_country,
			created_at, updated_at
		FROM customers WHERE email = ?
	`

	customer := &domain.Customer{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&customer.ID, &customer.Email, &customer.Name, &customer.Phone, &customer.LoyaltyPoints,
		&customer.Address.Street, &customer.Address.City, &customer.Address.State,
		&customer.Address.PostalCode, &customer.Address.Country,
		&customer.CreatedAt, &customer.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("customer")
	}

	return customer, err
}

func (r *SQLiteRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	query := `
		UPDATE customers SET email = ?, name = ?, phone = ?, loyalty_points = ?,
			address_street = ?, address_city = ?, address_state = ?, 
			address_postal_code = ?, address_country = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		customer.Email, customer.Name, customer.Phone, customer.LoyaltyPoints,
		customer.Address.Street, customer.Address.City, customer.Address.State,
		customer.Address.PostalCode, customer.Address.Country,
		time.Now(), customer.ID,
	)

	return err
}

func (r *SQLiteRepository) ListCustomers(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	query := `
		SELECT id, email, name, phone, loyalty_points,
			address_street, address_city, address_state, address_postal_code, address_country,
			created_at, updated_at
		FROM customers
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	customers := []*domain.Customer{}
	for rows.Next() {
		customer := &domain.Customer{}
		err := rows.Scan(
			&customer.ID, &customer.Email, &customer.Name, &customer.Phone, &customer.LoyaltyPoints,
			&customer.Address.Street, &customer.Address.City, &customer.Address.State,
			&customer.Address.PostalCode, &customer.Address.Country,
			&customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}

	return customers, nil
}

func (r *SQLiteRepository) CreateProduct(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (id, name, description, price, sku, stock, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		product.ID, product.Name, product.Description, product.Price,
		product.SKU, product.Stock, product.Category,
		product.CreatedAt, product.UpdatedAt,
	)

	return err
}

func (r *SQLiteRepository) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT id, name, description, price, sku, stock, category, created_at, updated_at FROM products WHERE id = ?`

	product := &domain.Product{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price,
		&product.SKU, &product.Stock, &product.Category,
		&product.CreatedAt, &product.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("product")
	}

	return product, err
}

func (r *SQLiteRepository) UpdateProduct(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products SET name = ?, description = ?, price = ?, stock = ?, category = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		product.Name, product.Description, product.Price, product.Stock,
		product.Category, time.Now(), product.ID,
	)

	return err
}

func (r *SQLiteRepository) ListProducts(ctx context.Context, limit, offset int) ([]*domain.Product, error) {
	query := `SELECT id, name, description, price, sku, stock, category, created_at, updated_at FROM products LIMIT ? OFFSET ?`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	products := []*domain.Product{}
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price,
			&product.SKU, &product.Stock, &product.Category,
			&product.CreatedAt, &product.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *SQLiteRepository) CreateCart(ctx context.Context, cart *domain.Cart) error {
	itemsJSON, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}

	query := `INSERT INTO carts (id, customer_id, items, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query, cart.ID, cart.CustomerID, string(itemsJSON), cart.CreatedAt, cart.UpdatedAt)
	return err
}

func (r *SQLiteRepository) GetCart(ctx context.Context, id string) (*domain.Cart, error) {
	query := `SELECT id, customer_id, items, created_at, updated_at FROM carts WHERE id = ?`

	var itemsJSON string
	cart := &domain.Cart{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cart.ID, &cart.CustomerID, &itemsJSON, &cart.CreatedAt, &cart.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("cart")
	}

	if err := json.Unmarshal([]byte(itemsJSON), &cart.Items); err != nil {
		return nil, err
	}

	return cart, err
}

func (r *SQLiteRepository) UpdateCart(ctx context.Context, cart *domain.Cart) error {
	itemsJSON, err := json.Marshal(cart.Items)
	if err != nil {
		return err
	}

	query := `UPDATE carts SET items = ?, updated_at = ? WHERE id = ?`
	_, err = r.db.ExecContext(ctx, query, string(itemsJSON), time.Now(), cart.ID)
	return err
}

func (r *SQLiteRepository) GetCartByCustomer(ctx context.Context, customerID string) (*domain.Cart, error) {
	query := `SELECT id, customer_id, items, created_at, updated_at FROM carts WHERE customer_id = ? ORDER BY updated_at DESC LIMIT 1`

	var itemsJSON string
	cart := &domain.Cart{}

	err := r.db.QueryRowContext(ctx, query, customerID).Scan(
		&cart.ID, &cart.CustomerID, &itemsJSON, &cart.CreatedAt, &cart.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("cart")
	}

	if err := json.Unmarshal([]byte(itemsJSON), &cart.Items); err != nil {
		return nil, err
	}

	return cart, err
}

func (r *SQLiteRepository) CreateTransaction(ctx context.Context, transaction *domain.Transaction) error {
	detailsJSON, _ := json.Marshal(transaction.PaymentDetails)
	metadataJSON, _ := json.Marshal(transaction.Metadata)

	query := `
		INSERT INTO transactions (id, customer_id, amount, status, payment_method, payment_details, metadata, error_message, processed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		transaction.ID, transaction.CustomerID, transaction.Amount, transaction.Status,
		transaction.PaymentMethod, string(detailsJSON), string(metadataJSON),
		transaction.ErrorMessage, transaction.ProcessedAt, transaction.CreatedAt,
	)

	return err
}

func (r *SQLiteRepository) GetTransaction(ctx context.Context, id string) (*domain.Transaction, error) {
	query := `SELECT id, customer_id, amount, status, payment_method, payment_details, metadata, error_message, processed_at, created_at FROM transactions WHERE id = ?`

	var detailsJSON, metadataJSON string
	transaction := &domain.Transaction{}

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&transaction.ID, &transaction.CustomerID, &transaction.Amount, &transaction.Status,
		&transaction.PaymentMethod, &detailsJSON, &metadataJSON,
		&transaction.ErrorMessage, &transaction.ProcessedAt, &transaction.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NewNotFoundError("transaction")
	}

	json.Unmarshal([]byte(detailsJSON), &transaction.PaymentDetails)
	json.Unmarshal([]byte(metadataJSON), &transaction.Metadata)

	return transaction, err
}

func (r *SQLiteRepository) ListTransactionsByCustomer(ctx context.Context, customerID string, limit, offset int) ([]*domain.Transaction, error) {
	query := `
		SELECT id, customer_id, amount, status, payment_method, payment_details, metadata, error_message, processed_at, created_at
		FROM transactions
		WHERE customer_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, customerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*domain.Transaction{}
	for rows.Next() {
		var detailsJSON, metadataJSON string
		transaction := &domain.Transaction{}

		err := rows.Scan(
			&transaction.ID, &transaction.CustomerID, &transaction.Amount, &transaction.Status,
			&transaction.PaymentMethod, &detailsJSON, &metadataJSON,
			&transaction.ErrorMessage, &transaction.ProcessedAt, &transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal([]byte(detailsJSON), &transaction.PaymentDetails)
		json.Unmarshal([]byte(metadataJSON), &transaction.Metadata)

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *SQLiteRepository) Close() error {
	return r.db.Close()
}
