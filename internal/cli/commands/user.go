package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ecommerce/payment-system/internal/domain"
	"github.com/ecommerce/payment-system/pkg/validator"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users/customers",
	Long:  `Register new users, view user information, and manage customer accounts.`,
}

var userRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new customer",
	Long:  `Register a new customer account with email, name, and contact information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		email, _ := cmd.Flags().GetString("email")
		name, _ := cmd.Flags().GetString("name")
		phone, _ := cmd.Flags().GetString("phone")
		street, _ := cmd.Flags().GetString("street")
		city, _ := cmd.Flags().GetString("city")
		state, _ := cmd.Flags().GetString("state")
		postalCode, _ := cmd.Flags().GetString("postal-code")
		country, _ := cmd.Flags().GetString("country")

		if email == "" || name == "" {
			return fmt.Errorf("email and name are required")
		}

		emailValidator := validator.NewEmailValidator()
		if err := emailValidator.Validate(email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}

		if phone != "" {
			phoneValidator := validator.NewPhoneValidator()
			if err := phoneValidator.Validate(phone); err != nil {
				return fmt.Errorf("invalid phone: %w", err)
			}
		}

		_, err := app.Repository.GetCustomerByEmail(ctx, email)
		if err == nil {
			color.Yellow("⚠ Customer with email %s already exists", email)
			return nil
		}

		customer := &domain.Customer{
			ID:            domain.NewID(),
			Email:         email,
			Name:          name,
			Phone:         phone,
			LoyaltyPoints: 0,
			Address: domain.Address{
				Street:     street,
				City:       city,
				State:      state,
				PostalCode: postalCode,
				Country:    country,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := app.Repository.CreateCustomer(ctx, customer); err != nil {
			return fmt.Errorf("failed to create customer: %w", err)
		}

		color.Green("\n✓ Customer registered successfully!")
		fmt.Printf("\nCustomer ID: %s\n", customer.ID)
		fmt.Printf("Email: %s\n", customer.Email)
		fmt.Printf("Name: %s\n", customer.Name)
		if customer.Phone != "" {
			fmt.Printf("Phone: %s\n", customer.Phone)
		}
		fmt.Printf("\n💡 Use this email to manage your cart and make purchases\n")
		fmt.Printf("   Set environment variable: export CUSTOMER_EMAIL=%s\n", customer.Email)

		return nil
	},
}

var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all customers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		customers, err := app.Repository.ListCustomers(ctx, 50, 0)
		if err != nil {
			return err
		}

		if len(customers) == 0 {
			fmt.Println("No customers found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Email", "Phone", "Loyalty Points", "State"})

		for _, customer := range customers {
			table.Append([]string{
				customer.ID[:8] + "...",
				customer.Name,
				customer.Email,
				customer.Phone,
				fmt.Sprintf("%d", customer.LoyaltyPoints),
				customer.Address.State,
			})
		}

		table.Render()
		fmt.Printf("\nTotal Customers: %d\n", len(customers))

		return nil
	},
}

var userInfoCmd = &cobra.Command{
	Use:   "info [email]",
	Short: "View customer information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		email := args[0]

		customer, err := app.Repository.GetCustomerByEmail(ctx, email)
		if err != nil {
			color.Red("✗ Customer not found: %s", email)
			return nil
		}

		color.Cyan("\n═══════════════════════════════════════")
		color.Cyan("          CUSTOMER INFORMATION")
		color.Cyan("═══════════════════════════════════════\n")

		fmt.Printf("Customer ID:    %s\n", customer.ID)
		fmt.Printf("Name:           %s\n", customer.Name)
		fmt.Printf("Email:          %s\n", customer.Email)
		if customer.Phone != "" {
			fmt.Printf("Phone:          %s\n", customer.Phone)
		}
		fmt.Printf("Loyalty Points: %d points\n", customer.LoyaltyPoints)
		fmt.Printf("Member Since:   %s\n", customer.CreatedAt.Format("2006-01-02"))

		if customer.Address.Street != "" {
			fmt.Printf("\nAddress:\n")
			fmt.Printf("  %s\n", customer.Address.Street)
			fmt.Printf("  %s, %s %s\n", customer.Address.City, customer.Address.State, customer.Address.PostalCode)
			fmt.Printf("  %s\n", customer.Address.Country)
		}

		color.Cyan("\n═══════════════════════════════════════\n")

		return nil
	},
}

func init() {
	userRegisterCmd.Flags().String("email", "", "Customer email (required)")
	userRegisterCmd.Flags().String("name", "", "Customer name (required)")
	userRegisterCmd.Flags().String("phone", "", "Customer phone number")
	userRegisterCmd.Flags().String("street", "", "Street address")
	userRegisterCmd.Flags().String("city", "", "City")
	userRegisterCmd.Flags().String("state", "", "State/Province")
	userRegisterCmd.Flags().String("postal-code", "", "Postal/ZIP code")
	userRegisterCmd.Flags().String("country", "USA", "Country")

	userCmd.AddCommand(userRegisterCmd)
	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userInfoCmd)
}
