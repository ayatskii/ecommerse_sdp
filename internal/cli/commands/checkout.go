package commands

import (
	"context"
	"fmt"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	paymentMethod     string
	paymentStrategy   string
	enabledDecorators []string
	discountCode      string
	useLoyaltyPoints  int
)

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Process checkout and payment",
	Long:  `Process checkout for the current cart with selected payment method and decorators.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return fmt.Errorf("failed to get customer: %w", err)
		}

		cart, err := app.CartService.GetOrCreateCart(ctx, customer.ID)
		if err != nil {
			return fmt.Errorf("failed to get cart: %w", err)
		}

		if len(cart.Items) == 0 {
			color.Yellow("⚠ Cart is empty. Add items first using 'cart add' command.")
			return nil
		}

		printCartSummary(cart)

		fmt.Println()
		color.Cyan("Customer Information:")
		fmt.Printf("  Name: %s\n", customer.Name)
		fmt.Printf("  Email: %s\n", customer.Email)
		fmt.Printf("  Loyalty Points: %d\n", customer.LoyaltyPoints)

		fmt.Println()
		color.Cyan("Payment Options:")
		fmt.Printf("  Payment Method: %s\n", paymentMethod)
		fmt.Printf("  Payment Strategy: %s\n", paymentStrategy)
		if len(enabledDecorators) > 0 {
			fmt.Printf("  Enabled Decorators: %v\n", enabledDecorators)
		}
		if discountCode != "" {
			fmt.Printf("  Discount Code: %s\n", discountCode)
		}
		if useLoyaltyPoints > 0 {
			fmt.Printf("  Using Loyalty Points: %d\n", useLoyaltyPoints)
		}

		options := domain.CheckoutOptions{
			PaymentMethod:     paymentMethod,
			PaymentStrategy:   paymentStrategy,
			EnabledDecorators: enabledDecorators,
			DiscountCode:      discountCode,
			UseLoyaltyPoints:  useLoyaltyPoints,
		}

		fmt.Println()
		color.Yellow("⏳ Processing checkout...")

		receipt, err := app.CheckoutFacade.ProcessOrder(ctx, cart, customer, options)
		if err != nil {
			color.Red("✗ Checkout failed: %v", err)
			return nil
		}

		fmt.Println()
		printReceipt(receipt)

		color.Green("✓ Checkout completed successfully!")

		return nil
	},
}

func init() {
	checkoutCmd.Flags().StringVarP(&paymentMethod, "method", "m", "credit_card", "Payment method (credit_card, paypal, crypto)")
	checkoutCmd.Flags().StringVarP(&paymentStrategy, "strategy", "s", "instant", "Payment strategy (instant, deferred, split)")
	checkoutCmd.Flags().StringSliceVarP(&enabledDecorators, "decorators", "d", []string{"tax", "fraud_detection"}, "Enabled decorators")
	checkoutCmd.Flags().StringVar(&discountCode, "discount", "", "Discount code")
	checkoutCmd.Flags().IntVarP(&useLoyaltyPoints, "points", "p", 0, "Loyalty points to use")
}

func printCartSummary(cart *domain.Cart) {
	color.Cyan("Cart Summary:")
	fmt.Printf("  Items: %d\n", cart.GetItemCount())
	fmt.Printf("  Total: $%.2f\n", cart.GetTotal())
	fmt.Println()
	fmt.Println("  Items:")
	for _, item := range cart.Items {
		fmt.Printf("    - %s x%d @ $%.2f = $%.2f\n",
			item.Product.Name,
			item.Quantity,
			item.Price,
			item.Price*float64(item.Quantity),
		)
	}
}

func printReceipt(receipt *domain.Receipt) {
	color.Cyan("═══════════════════════════════════════")
	color.Cyan("              RECEIPT")
	color.Cyan("═══════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Transaction ID: %s\n", receipt.TransactionID)
	fmt.Printf("Date: %s\n", receipt.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println()

	fmt.Printf("Customer: %s\n", receipt.CustomerName)
	fmt.Printf("Email: %s\n", receipt.CustomerEmail)
	fmt.Println()

	color.Cyan("Items:")
	for _, item := range receipt.Items {
		fmt.Printf("  %-30s x%-3d $%8.2f\n",
			item.ProductName,
			item.Quantity,
			item.Total,
		)
	}
	fmt.Println()

	color.Cyan("Amounts:")
	fmt.Printf("  Subtotal:          $%8.2f\n", receipt.Subtotal)
	if receipt.Discount > 0 {
		fmt.Printf("  Discount:          -$%8.2f\n", receipt.Discount)
	}
	if receipt.Tax > 0 {
		fmt.Printf("  Tax:               $%8.2f\n", receipt.Tax)
	}
	color.Green("  Total:             $%8.2f\n", receipt.Total)
	fmt.Println()

	if receipt.Cashback > 0 {
		color.Yellow("  Cashback Earned:   $%8.2f\n", receipt.Cashback)
	}
	if receipt.LoyaltyPoints > 0 {
		color.Yellow("  Loyalty Points:    %d points\n", receipt.LoyaltyPoints)
	}

	if len(receipt.AppliedDecorators) > 0 {
		fmt.Println()
		fmt.Printf("Applied Features: %v\n", receipt.AppliedDecorators)
	}

	fmt.Println()
	color.Cyan("═══════════════════════════════════════")
}
