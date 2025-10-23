package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var exchangeRates = map[string]float64{
	"USD": 538.0,
	"EUR": 580.0,
	"RUB": 5.8,
	"CNY": 75.0,
	"KZT": 1.0,
}

var (
	fromCurrency string
	toCurrency   string
	number       int
)

var debitCmd = &cobra.Command{
	Use:   "debit",
	Short: "Debit cart and convert total to specified currency",
	Long:  `Process payment for cart total and convert the amount to specified currency.`,
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
			color.Yellow("  Cart is empty. Add items first using 'cart add' command.")
			return nil
		}

		originalAmount := cart.GetTotal()
		convertedAmount := convertCurrency(originalAmount, fromCurrency, toCurrency)

		color.Cyan("Cart Summary:")
		fmt.Printf("  Items: %d\n", cart.GetItemCount())
		fmt.Printf("  Total (%s): %.2f %s\n", fromCurrency, originalAmount, fromCurrency)
		if fromCurrency != toCurrency {
			var rate float64
			if toCurrency != "" {
				rate = exchangeRates[fromCurrency] / exchangeRates[toCurrency]
			} else {
				rate = exchangeRates[fromCurrency]
			}
			fmt.Printf("  Exchange Rate: 1 %s = %.4f %s\n", fromCurrency, rate, toCurrency)
		}
		color.Green("  Total (%s): %.2f %s\n", toCurrency, convertedAmount, toCurrency)

		fmt.Println()
		transaction := &domain.Transaction{
			ID:            domain.NewID(),
			CustomerID:    customer.ID,
			Amount:        convertedAmount,
			Status:        domain.TransactionStatusCompleted,
			PaymentMethod: fmt.Sprintf("debit_%s", strings.ToLower(toCurrency)),
			PaymentDetails: map[string]interface{}{
				"original_amount":    originalAmount,
				"original_currency":  fromCurrency,
				"converted_amount":   convertedAmount,
				"converted_currency": toCurrency,
			},
			ProcessedAt: time.Now(),
			CreatedAt:   time.Now(),
		}
		color.Green("  Debit payment processed successfully!")
		fmt.Printf("  Transaction ID: %s\n", transaction.ID)
		amoundDebited := convertCurrency(float64(number), fromCurrency, toCurrency)
		fmt.Printf("  Amount debited: %.2f %s\n", amoundDebited, toCurrency)
		if amoundDebited < convertedAmount {
			color.Red("  Insufficient fund")
			fmt.Println()
		} else {
			color.Green("  Payment proceeded succesfully")
			remainingAmount := amoundDebited - convertedAmount
			fmt.Printf("  Remaining on balance %.2f", remainingAmount)
			fmt.Println()
		}

		if err := app.CartService.ClearCart(ctx, cart.ID); err != nil {
			color.Yellow("  Failed to clear cart: %v", err)
		} else {
			color.Green("  Cart cleared after successful payment")
		}

		return nil
	},
}

func convertCurrency(amount float64, from, to string) float64 {
	from = strings.ToUpper(from)
	to = strings.ToUpper(to)

	if from == to {
		return amount
	}

	amountInKZT := amount * exchangeRates[from]
	return amountInKZT / exchangeRates[to]
}

func init() {
	debitCmd.Flags().StringVarP(&fromCurrency, "from", "f", "USD", "Source currency")
	debitCmd.Flags().StringVarP(&toCurrency, "to", "t", "KZT", "Target currency")
	debitCmd.Flags().IntVarP(&number, "number", "n", 1000, "Target number")
}
