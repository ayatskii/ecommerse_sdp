package commands

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ecommerce/payment-system/internal/app"
	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func getCustomer(ctx context.Context, application *app.Application) (*domain.Customer, error) {
	email := os.Getenv("CUSTOMER_EMAIL")
	if email == "" {
		email = "john.doe@example.com"
	}

	customer, err := application.Repository.GetCustomerByEmail(ctx, email)
	if err != nil {
		color.Yellow("⚠ Customer not found. Please register first:")
		color.Yellow("  bin/ecommerce-cli.exe user register --email your@email.com --name \"Your Name\"")
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

var cartCmd = &cobra.Command{
	Use:   "cart",
	Short: "Manage shopping cart",
	Long:  `Add, remove, or view items in the shopping cart.`,
}

var cartViewCmd = &cobra.Command{
	Use:   "view",
	Short: "View cart contents",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return err
		}

		cart, err := app.CartService.GetOrCreateCart(ctx, customer.ID)
		if err != nil {
			return err
		}

		if len(cart.Items) == 0 {
			color.Yellow("Cart is empty")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Product", "SKU", "Price", "Quantity", "Total"})

		for _, item := range cart.Items {
			table.Append([]string{
				item.Product.Name,
				item.Product.SKU,
				fmt.Sprintf("$%.2f", item.Price),
				fmt.Sprintf("%d", item.Quantity),
				fmt.Sprintf("$%.2f", item.Price*float64(item.Quantity)),
			})
		}

		table.SetFooter([]string{"", "", "", "Total", fmt.Sprintf("$%.2f", cart.GetTotal())})
		table.Render()

		return nil
	},
}

var cartAddCmd = &cobra.Command{
	Use:   "add [product-id] [quantity]",
	Short: "Add item to cart",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		productID := args[0]
		quantity, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid quantity: %w", err)
		}

		product, err := app.Repository.GetProduct(ctx, productID)
		if err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return err
		}

		cart, err := app.CartService.GetOrCreateCart(ctx, customer.ID)
		if err != nil {
			return err
		}

		if err := app.CartService.AddItem(ctx, cart.ID, product, quantity); err != nil {
			return err
		}

		color.Green("✓ Added %s x%d to cart", product.Name, quantity)
		return nil
	},
}

var cartRemoveCmd = &cobra.Command{
	Use:   "remove [product-id]",
	Short: "Remove item from cart",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		productID := args[0]

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return err
		}

		cart, err := app.CartService.GetOrCreateCart(ctx, customer.ID)
		if err != nil {
			return err
		}

		if err := app.CartService.RemoveItem(ctx, cart.ID, productID); err != nil {
			return err
		}

		color.Green("✓ Item removed from cart")
		return nil
	},
}

var cartClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all items from cart",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return err
		}

		cart, err := app.CartService.GetOrCreateCart(ctx, customer.ID)
		if err != nil {
			return err
		}

		if err := app.CartService.ClearCart(ctx, cart.ID); err != nil {
			return err
		}

		color.Green("✓ Cart cleared")
		return nil
	},
}

func init() {
	cartCmd.AddCommand(cartViewCmd)
	cartCmd.AddCommand(cartAddCmd)
	cartCmd.AddCommand(cartRemoveCmd)
	cartCmd.AddCommand(cartClearCmd)
}
