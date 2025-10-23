package domain

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	LoyaltyPoints int       `json:"loyalty_points"`
	Address       Address   `json:"address"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	SKU         string    `json:"sku"`
	Stock       int       `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItem struct {
	ProductID string  `json:"product_id"`
	Product   Product `json:"product"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type Cart struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	Items      []CartItem `json:"items"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (c *Cart) GetTotal() float64 {
	total := 0.0
	for _, item := range c.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}

func (c *Cart) GetItemCount() int {
	count := 0
	for _, item := range c.Items {
		count += item.Quantity
	}
	return count
}

func (c *Cart) AddItem(product Product, quantity int) {

	for i, item := range c.Items {
		if item.ProductID == product.ID {
			c.Items[i].Quantity += quantity
			return
		}
	}

	c.Items = append(c.Items, CartItem{
		ProductID: product.ID,
		Product:   product,
		Quantity:  quantity,
		Price:     product.Price,
	})
}

func (c *Cart) RemoveItem(productID string) {
	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return
		}
	}
}

func (c *Cart) UpdateQuantity(productID string, quantity int) {
	if quantity <= 0 {
		c.RemoveItem(productID)
		return
	}

	for i, item := range c.Items {
		if item.ProductID == productID {
			c.Items[i].Quantity = quantity
			return
		}
	}
}

func (c *Cart) Clear() {
	c.Items = []CartItem{}
}

type Transaction struct {
	ID             string                 `json:"id"`
	CustomerID     string                 `json:"customer_id"`
	Amount         float64                `json:"amount"`
	Status         TransactionStatus      `json:"status"`
	PaymentMethod  string                 `json:"payment_method"`
	PaymentDetails map[string]interface{} `json:"payment_details"`
	Metadata       map[string]interface{} `json:"metadata"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	ProcessedAt    time.Time              `json:"processed_at"`
	CreatedAt      time.Time              `json:"created_at"`
}

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusRefunded   TransactionStatus = "refunded"
)

type Receipt struct {
	ID                string                 `json:"id"`
	TransactionID     string                 `json:"transaction_id"`
	CustomerID        string                 `json:"customer_id"`
	CustomerName      string                 `json:"customer_name"`
	CustomerEmail     string                 `json:"customer_email"`
	Items             []ReceiptItem          `json:"items"`
	Subtotal          float64                `json:"subtotal"`
	Discount          float64                `json:"discount"`
	Tax               float64                `json:"tax"`
	Cashback          float64                `json:"cashback"`
	LoyaltyPoints     int                    `json:"loyalty_points_earned"`
	Total             float64                `json:"total"`
	PaymentMethod     string                 `json:"payment_method"`
	PaymentDetails    map[string]interface{} `json:"payment_details"`
	AppliedDecorators []string               `json:"applied_decorators"`
	CreatedAt         time.Time              `json:"created_at"`
}

type ReceiptItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

type Discount struct {
	ID          string       `json:"id"`
	Code        string       `json:"code"`
	Description string       `json:"description"`
	Type        DiscountType `json:"type"`
	Value       float64      `json:"value"`
	MinAmount   float64      `json:"min_amount"`
	MaxAmount   float64      `json:"max_amount"`
	ExpiresAt   time.Time    `json:"expires_at"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
}

type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
)

func (d *Discount) IsValid() bool {
	if !d.IsActive {
		return false
	}
	if !d.ExpiresAt.IsZero() && time.Now().After(d.ExpiresAt) {
		return false
	}
	return true
}

func (d *Discount) Calculate(amount float64) float64 {
	if !d.IsValid() {
		return 0
	}

	if amount < d.MinAmount {
		return 0
	}

	var discountAmount float64
	if d.Type == DiscountTypePercentage {
		discountAmount = amount * (d.Value / 100.0)
	} else {
		discountAmount = d.Value
	}

	if d.MaxAmount > 0 && discountAmount > d.MaxAmount {
		discountAmount = d.MaxAmount
	}

	return discountAmount
}

type CheckoutOptions struct {
	PaymentMethod     string                 `json:"payment_method"`
	PaymentStrategy   string                 `json:"payment_strategy"`
	EnabledDecorators []string               `json:"enabled_decorators"`
	DiscountCode      string                 `json:"discount_code,omitempty"`
	UseLoyaltyPoints  int                    `json:"use_loyalty_points,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

func NewID() string {
	return uuid.New().String()
}
