package cmd

import (
	"os"
	"ecommerce-cli/internal/app"
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Lance l'interface graphique interactive (Charmbracelet)",
	Run: func(cmd *cobra.Command, args []string) {
		app.RunClientApp(BaseURL, os.Stdin, os.Stdout)
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
