package cmd

import (
	"fmt"
	"os"

	"ecommerce-cli/internal/app"

	"github.com/spf13/cobra"
)

const BaseURL = "http://localhost:8080"

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "E-Commerce CLI Client",
	Long:  `A modern CLI application for the E-Commerce platform using Cobra and Charmbracelet.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Par défaut, sans argument, on lance le menu interactif (Charmbracelet)
		app.RunClientApp(BaseURL, os.Stdin, os.Stdout)
	},
}

// Execute is the entry point for Cobra commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
