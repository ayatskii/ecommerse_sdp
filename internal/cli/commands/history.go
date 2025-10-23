package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View transaction history",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		customer, err := getCustomer(ctx, app)
		if err != nil {
			return err
		}

		transactionService := app.Repository
		transactions, err := transactionService.ListTransactionsByCustomer(ctx, customer.ID, 50, 0)
		if err != nil {
			return err
		}

		if len(transactions) == 0 {
			fmt.Println("No transaction history found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Transaction ID", "Amount", "Method", "Status", "Date"})

		for _, tx := range transactions {
			table.Append([]string{
				tx.ID[:8] + "...",
				fmt.Sprintf("$%.2f", tx.Amount),
				tx.PaymentMethod,
				string(tx.Status),
				tx.CreatedAt.Format("2006-01-02 15:04"),
			})
		}

		table.Render()

		fmt.Printf("\nTotal Transactions: %d\n", len(transactions))

		return nil
	},
}
