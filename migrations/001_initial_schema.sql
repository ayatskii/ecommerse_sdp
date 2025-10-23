-- E-Commerce Payment System Database Schema

-- Customers table
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

-- Products table
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

-- Carts table
CREATE TABLE IF NOT EXISTS carts (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);

-- Cart items table
CREATE TABLE IF NOT EXISTS cart_items (
    id TEXT PRIMARY KEY,
    cart_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    price REAL NOT NULL,
    FOREIGN KEY (cart_id) REFERENCES carts(id),
    FOREIGN KEY (product_id) REFERENCES products(id)
);

-- Transactions table
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

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_carts_customer ON carts(customer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_customer ON transactions(customer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

