package orders

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ecommerce-cli/internal/models"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Margin(1, 0)
	orderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1).MarginBottom(1).BorderForeground(lipgloss.Color("63"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	infoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
)

type OrderClient struct {
	BaseURL string
	Token   string
}

func NewOrderClient(baseURL, token string) *OrderClient {
	return &OrderClient{
		BaseURL: baseURL,
		Token:   token,
	}
}

func (c *OrderClient) doAuthRequest(method, path string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	return http.DefaultClient.Do(req)
}

func (c *OrderClient) fetchOrders() ([]models.Order, error) {
	resp, err := c.doAuthRequest(http.MethodGet, "/api/orders")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(string(body))
	}

	var orders []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (c *OrderClient) RunOrdersMenu() {
	for {
		fmt.Println(titleStyle.Render("=== Mes Commandes ==="))
		
		orders, err := c.fetchOrders()
		if err != nil {
			fmt.Println(errorStyle.Render("Erreur: Impossible de récupérer vos commandes."))
			return
		}

		if len(orders) == 0 {
			fmt.Println(infoStyle.Render("Vous n'avez passé aucune commande pour le moment."))
		} else {
			for _, o := range orders {
				dateStr := o.CreatedAt.Format("02/01/2006 à 15:04")
				
				var itemsStr []string
				for _, item := range o.Items {
					prodName := "Produit inconnu"
					if item.Product != nil {
						prodName = item.Product.Name
					}
					itemsStr = append(itemsStr, fmt.Sprintf("  - %s (x%d) : %.2f €", prodName, item.Quantity, item.UnitPrice))
				}
				
				content := fmt.Sprintf("Commande #%d (Identifiant: %s)\nDate: %s\nStatut: %s\nTotal TTC: %.2f €\n\nProduits:\n%s",
					o.ID, o.BusinessID, dateStr, string(o.Status), o.TotalTTC, strings.Join(itemsStr, "\n"))
					
				if o.CancelReason != "" {
					content += fmt.Sprintf("\n\nRaison (si annulée): %s", o.CancelReason)
				}
				fmt.Println(orderStyle.Render(content))
			}
		}

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Action").
					Options(
						huh.NewOption("Retour au menu principal", "back"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil || action == "back" {
			return
		}
	}
}
