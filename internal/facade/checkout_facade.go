package facade

import (
	"context"
	"fmt"
	"time"

	"github.com/ecommerce/payment-system/config"
	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/internal/factory"
	"github.com/ecommerce/payment-system/internal/observer"
	"github.com/ecommerce/payment-system/internal/payment"
	"github.com/ecommerce/payment-system/internal/repository"
	"github.com/ecommerce/payment-system/internal/service"
	"github.com/ecommerce/payment-system/internal/strategy"
	"github.com/ecommerce/payment-system/pkg/errors"
	"github.com/ecommerce/payment-system/pkg/logger"
	"go.uber.org/zap"
)

type CheckoutFacade struct {
	config             *config.Config
	paymentFactory     *factory.PaymentFactory
	decoratorFactory   *factory.DecoratorFactory
	strategyFactory    *factory.StrategyFactory
	inventoryService   *service.InventoryService
	customerService    *service.CustomerService
	transactionService *service.TransactionService
	eventSubject       *observer.Subject
}

func NewCheckoutFacade(
	cfg *config.Config,
	repo repository.Repository,
	eventSubject *observer.Subject,
) *CheckoutFacade {
	return &CheckoutFacade{
		config:             cfg,
		paymentFactory:     factory.NewPaymentFactory(),
		decoratorFactory:   factory.NewDecoratorFactory(cfg),
		strategyFactory:    factory.NewStrategyFactory(),
		inventoryService:   service.NewInventoryService(repo),
		customerService:    service.NewCustomerService(repo),
		transactionService: service.NewTransactionService(repo),
		eventSubject:       eventSubject,
	}
}

func (f *CheckoutFacade) ProcessOrder(
	ctx context.Context,
	cart *domain.Cart,
	customer *domain.Customer,
	options domain.CheckoutOptions,
) (*domain.Receipt, error) {
	logger.Info("Starting checkout process",
		zap.String("customer_id", customer.ID),
		zap.String("cart_id", cart.ID),
		zap.Float64("amount", cart.GetTotal()),
	)

	transaction := &domain.Transaction{
		ID:             domain.NewID(),
		CustomerID:     customer.ID,
		Amount:         cart.GetTotal(),
		Status:         domain.TransactionStatusPending,
		PaymentMethod:  options.PaymentMethod,
		PaymentDetails: make(map[string]interface{}),
		Metadata:       options.Metadata,
		CreatedAt:      time.Now(),
	}

	f.notifyEvent(ctx, observer.Event{
		Type:          observer.EventPaymentStarted,
		TransactionID: transaction.ID,
		CustomerID:    customer.ID,
		Amount:        cart.GetTotal(),
		PaymentMethod: options.PaymentMethod,
		Timestamp:     time.Now().Format(time.RFC3339),
	})

	if err := f.validateInventory(ctx, cart); err != nil {
		return nil, f.handleError(ctx, transaction, err, "inventory validation failed")
	}

	if err := f.reserveInventory(ctx, cart); err != nil {
		return nil, f.handleError(ctx, transaction, err, "inventory reservation failed")
	}

	paymentInstance, err := f.createPayment(options)
	if err != nil {
		f.rollbackInventory(ctx, cart)
		return nil, f.handleError(ctx, transaction, err, "payment creation failed")
	}

	decoratedPayment, err := f.applyDecorators(ctx, paymentInstance, options, customer)
	if err != nil {
		f.rollbackInventory(ctx, cart)
		return nil, f.handleError(ctx, transaction, err, "decorator application failed")
	}

	result, err := f.executePaymentStrategy(ctx, decoratedPayment, cart.GetTotal(), options)
	if err != nil {
		f.rollbackInventory(ctx, cart)
		return nil, f.handleError(ctx, transaction, err, "payment processing failed")
	}

	transaction.Status = domain.TransactionStatusCompleted
	transaction.ProcessedAt = time.Now()
	transaction.PaymentDetails = result.Metadata

	if err := f.updateLoyaltyPoints(ctx, customer, result); err != nil {
		logger.Warn("Failed to update loyalty points",
			zap.Error(err),
			zap.String("customer_id", customer.ID),
		)
	}

	receipt := f.generateReceipt(transaction, cart, customer, result)

	if err := f.transactionService.CreateTransaction(ctx, transaction); err != nil {
		logger.Error("Failed to save transaction",
			zap.Error(err),
			zap.String("transaction_id", transaction.ID),
		)
	}

	cart.Clear()

	f.notifyEvent(ctx, observer.Event{
		Type:          observer.EventPaymentSuccess,
		TransactionID: transaction.ID,
		CustomerID:    customer.ID,
		Amount:        result.Amount,
		PaymentMethod: result.PaymentMethod,
		Result:        result,
		Timestamp:     time.Now().Format(time.RFC3339),
	})

	logger.Info("Checkout completed successfully",
		zap.String("transaction_id", transaction.ID),
		zap.Float64("amount", result.Amount),
	)

	return receipt, nil
}

func (f *CheckoutFacade) validateInventory(ctx context.Context, cart *domain.Cart) error {
	logger.Debug("Validating inventory")

	for _, item := range cart.Items {
		available, err := f.inventoryService.CheckAvailability(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return errors.Wrap(err, errors.ErrCodeInventoryError, "failed to check inventory")
		}

		if !available {
			return errors.NewInventoryError(
				fmt.Sprintf("insufficient inventory for product %s", item.Product.Name),
			)
		}
	}

	return nil
}

func (f *CheckoutFacade) reserveInventory(ctx context.Context, cart *domain.Cart) error {
	logger.Debug("Reserving inventory")

	for _, item := range cart.Items {
		if err := f.inventoryService.ReserveStock(ctx, item.ProductID, item.Quantity); err != nil {
			return errors.Wrap(err, errors.ErrCodeInventoryError, "failed to reserve inventory")
		}
	}

	return nil
}

func (f *CheckoutFacade) rollbackInventory(ctx context.Context, cart *domain.Cart) {
	logger.Warn("Rolling back inventory reservations")

	for _, item := range cart.Items {
		if err := f.inventoryService.ReleaseStock(ctx, item.ProductID, item.Quantity); err != nil {
			logger.Error("Failed to rollback inventory",
				zap.Error(err),
				zap.String("product_id", item.ProductID),
			)
		}
	}
}

func (f *CheckoutFacade) createPayment(options domain.CheckoutOptions) (payment.Payment, error) {
	logger.Debug("Creating payment instance",
		zap.String("payment_method", options.PaymentMethod),
	)

	config := payment.PaymentConfig{}

	switch options.PaymentMethod {
	case "credit_card":
		config.CardNumber = "4532015112830366"
		config.CardHolder = "John Doe"
		config.ExpiryDate = "12/25"
		config.CVV = "123"
	case "paypal":
		config.PayPalEmail = "user@example.com"
		config.PayPalPassword = "password"
	case "crypto":
		config.WalletAddress = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"
		config.CryptoType = "BTC"
	}

	return f.paymentFactory.CreatePayment(options.PaymentMethod, config)
}

func (f *CheckoutFacade) applyDecorators(
	ctx context.Context,
	paymentInstance payment.Payment,
	options domain.CheckoutOptions,
	customer *domain.Customer,
) (payment.Payment, error) {
	logger.Debug("Applying decorators",
		zap.Strings("decorators", options.EnabledDecorators),
	)

	return f.decoratorFactory.CreateDecoratorChain(
		paymentInstance,
		options.EnabledDecorators,
		options,
		customer,
	)
}

func (f *CheckoutFacade) executePaymentStrategy(
	ctx context.Context,
	paymentInstance payment.Payment,
	amount float64,
	options domain.CheckoutOptions,
) (*payment.PaymentResult, error) {
	logger.Debug("Executing payment strategy",
		zap.String("strategy", options.PaymentStrategy),
		zap.Float64("amount", amount),
	)

	ctx, cancel := context.WithTimeout(ctx, f.config.Payment.Timeout)
	defer cancel()

	strategyType := options.PaymentStrategy
	if strategyType == "" {
		strategyType = "instant"
	}

	paymentStrategy, err := f.strategyFactory.CreateStrategy(strategyType, nil)
	if err != nil {
		return nil, err
	}

	return f.executeWithRetry(ctx, paymentStrategy, paymentInstance, amount)
}

func (f *CheckoutFacade) executeWithRetry(
	ctx context.Context,
	paymentStrategy strategy.PaymentStrategy,
	paymentInstance payment.Payment,
	amount float64,
) (*payment.PaymentResult, error) {
	var lastErr error

	for attempt := 0; attempt <= f.config.Payment.RetryAttempts; attempt++ {
		if attempt > 0 {
			logger.Info("Retrying payment",
				zap.Int("attempt", attempt),
			)
			time.Sleep(f.config.Payment.RetryDelay)
		}

		result, err := paymentStrategy.Execute(ctx, paymentInstance, amount)
		if err == nil {
			return result, nil
		}

		lastErr = err

		if errors.IsErrorCode(err, errors.ErrCodeFraudDetected) ||
			errors.IsErrorCode(err, errors.ErrCodeInvalidPayment) {
			break
		}
	}

	return nil, lastErr
}

func (f *CheckoutFacade) updateLoyaltyPoints(
	ctx context.Context,
	customer *domain.Customer,
	result *payment.PaymentResult,
) error {

	pointsEarned := 0
	if val, ok := result.Metadata["loyalty_points_earned"].(int); ok {
		pointsEarned = val
	}

	pointsRedeemed := 0
	if val, ok := result.Metadata["loyalty_points_redeemed"].(int); ok {
		pointsRedeemed = val
	}

	if pointsEarned > 0 || pointsRedeemed > 0 {
		return f.customerService.UpdateLoyaltyPoints(
			ctx,
			customer.ID,
			pointsEarned,
			pointsRedeemed,
		)
	}

	return nil
}

func (f *CheckoutFacade) generateReceipt(
	transaction *domain.Transaction,
	cart *domain.Cart,
	customer *domain.Customer,
	result *payment.PaymentResult,
) *domain.Receipt {

	items := make([]domain.ReceiptItem, 0, len(cart.Items))
	for _, item := range cart.Items {
		items = append(items, domain.ReceiptItem{
			ProductID:   item.ProductID,
			ProductName: item.Product.Name,
			SKU:         item.Product.SKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.Price,
			Total:       item.Price * float64(item.Quantity),
		})
	}

	subtotal := cart.GetTotal()
	discount := 0.0
	tax := 0.0
	cashback := 0.0
	loyaltyPoints := 0

	if val, ok := result.Metadata["discount_amount"].(float64); ok {
		discount = val
	}
	if val, ok := result.Metadata["tax_amount"].(float64); ok {
		tax = val
	}
	if val, ok := result.Metadata["cashback_amount"].(float64); ok {
		cashback = val
	}
	if val, ok := result.Metadata["loyalty_points_earned"].(int); ok {
		loyaltyPoints = val
	}

	return &domain.Receipt{
		ID:                domain.NewID(),
		TransactionID:     transaction.ID,
		CustomerID:        customer.ID,
		CustomerName:      customer.Name,
		CustomerEmail:     customer.Email,
		Items:             items,
		Subtotal:          subtotal,
		Discount:          discount,
		Tax:               tax,
		Cashback:          cashback,
		LoyaltyPoints:     loyaltyPoints,
		Total:             result.Amount,
		PaymentMethod:     result.PaymentMethod,
		PaymentDetails:    result.Metadata,
		AppliedDecorators: result.AppliedDecorators,
		CreatedAt:         time.Now(),
	}
}

func (f *CheckoutFacade) handleError(
	ctx context.Context,
	transaction *domain.Transaction,
	err error,
	message string,
) error {
	logger.Error(message,
		zap.Error(err),
		zap.String("transaction_id", transaction.ID),
	)

	transaction.Status = domain.TransactionStatusFailed
	transaction.ErrorMessage = err.Error()

	f.notifyEvent(ctx, observer.Event{
		Type:          observer.EventPaymentFailed,
		TransactionID: transaction.ID,
		CustomerID:    transaction.CustomerID,
		Amount:        transaction.Amount,
		PaymentMethod: transaction.PaymentMethod,
		Error:         err,
		Timestamp:     time.Now().Format(time.RFC3339),
	})

	return errors.Wrap(err, errors.ErrCodePaymentFailed, message)
}

func (f *CheckoutFacade) notifyEvent(ctx context.Context, event observer.Event) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Event notification panic",
					zap.Any("panic", r),
				)
			}
		}()

		f.eventSubject.Notify(context.Background(), event)
	}()
}
