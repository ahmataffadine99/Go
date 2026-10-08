package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Vérifie l'état du serveur Backend",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Vérification du serveur Backend...")
		
		client := http.Client{
			Timeout: 3 * time.Second,
		}
		
		resp, err := client.Get(BaseURL + "/api/products")
		if err != nil {
			fmt.Printf("❌ Serveur injoignable: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅ Serveur en ligne et opérationnel !")
		} else {
			fmt.Printf("⚠️ Serveur en ligne mais statut inattendu: %d\n", resp.StatusCode)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
