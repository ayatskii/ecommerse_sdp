package commands

import (
	"fmt"
	"os"

	"github.com/ecommerce/payment-system/internal/app"
	"github.com/spf13/cobra"
)

var (
	configPath  string
	application *app.Application
)

var rootCmd = &cobra.Command{
	Use:   "ecommerce-cli",
	Short: "E-Commerce Payment & Discount System",
	Long: `A production-grade CLI application for e-commerce payment processing
featuring multiple design patterns including Decorator, Strategy, Observer,
Factory, and Facade patterns.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {

		var err error
		application, err = app.Initialize(configPath)
		if err != nil {
			return fmt.Errorf("failed to initialize application: %w", err)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {

		if application != nil {
			return application.Shutdown()
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "./config", "config file directory")

	rootCmd.AddCommand(checkoutCmd)
	rootCmd.AddCommand(cartCmd)
	rootCmd.AddCommand(productsCmd)
	rootCmd.AddCommand(historyCmd)
	rootCmd.AddCommand(userCmd)
}

func GetApplication() *app.Application {
	return application
}

func exitOnError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
