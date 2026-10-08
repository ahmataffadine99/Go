package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Affiche la version de l'application",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("E-Commerce CLI Client v1.0.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
