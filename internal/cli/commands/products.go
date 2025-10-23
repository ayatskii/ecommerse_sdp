package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var productsCmd = &cobra.Command{
	Use:   "products",
	Short: "List available products",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		app := GetApplication()

		products, err := app.Repository.ListProducts(ctx, 100, 0)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "SKU", "Price", "Stock", "Category"})

		for _, product := range products {
			table.Append([]string{
				product.ID,
				product.Name,
				product.SKU,
				fmt.Sprintf("$%.2f", product.Price),
				fmt.Sprintf("%d", product.Stock),
				product.Category,
			})
		}

		table.Render()

		fmt.Printf("\nTotal Products: %d\n", len(products))

		return nil
	},
}
