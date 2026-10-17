package orders

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
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
	In      io.Reader
	Out     io.Writer
}

func NewOrderClient(baseURL, token string, in io.Reader, out io.Writer) *OrderClient {
	return &OrderClient{
		BaseURL: baseURL,
		Token:   token,
		In:      in,
		Out:     out,
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
		return nil, errors.New(string(body))
	}

	var orders []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func formatDate(s string) string {
	s = strings.Replace(s, "T", " ", 1)
	if len(s) > 16 {
		return s[:16]
	}
	return s
}

func (c *OrderClient) RunOrdersMenu() {
	var orders []models.Order
	err := components.RunWithSpinner(c.In, c.Out, "Chargement de vos commandes...", func() error {
		var err error
		orders, err = c.fetchOrders()
		return err
	})
	if err != nil {
		fmt.Fprintln(c.Out, errorStyle.Render("Erreur: Impossible de récupérer vos commandes."))
		return
	}

	if len(orders) == 0 {
		fmt.Fprintln(c.Out, titleStyle.Render("=== Mes Commandes ==="))
		fmt.Fprintln(c.Out, infoStyle.Render("Vous n'avez passé aucune commande pour le moment."))
		return
	}

	rows := make([][]string, len(orders))
	for i, o := range orders {
		status := string(o.Status)
		if o.Status == models.OrderStatusCanceled && o.CancelReason != "" {
			status += " (" + o.CancelReason + ")"
		}
		rows[i] = []string{o.BusinessID, formatDate(o.CreatedAt), fmt.Sprintf("%d", len(o.Items)), fmt.Sprintf("%.2f €", o.TotalTTC), status}
	}

	for {
		idx, ok := components.RunTable(c.In, c.Out, "Mes Commandes", []string{"Identifiant", "Date", "Produits", "Total TTC", "Statut"}, rows)
		if !ok {
			return
		}
		c.printOrder(orders[idx])

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Action").
					Options(
						huh.NewOption("Revenir à la liste", "list"),
						huh.NewOption("Retour au menu principal", "back"),
					).
					Value(&action),
			),
		)

		if err := form.WithProgramOptions(tea.WithInput(c.In), tea.WithOutput(c.Out)).Run(); err != nil || action == "back" {
			return
		}
	}
}

func (c *OrderClient) printOrder(o models.Order) {
	var itemsStr []string
	for _, item := range o.Items {
		prodName := "Produit inconnu"
		if item.Product != nil {
			prodName = item.Product.Name
		}
		itemsStr = append(itemsStr, fmt.Sprintf("  - %s (x%d) : %.2f € TTC l'unité", prodName, item.Quantity, item.UnitPrice))
	}

	content := fmt.Sprintf("Commande %s\nDate: %s\nStatut: %s\nTotal TTC: %.2f €\n\nProduits:\n%s",
		o.BusinessID, formatDate(o.CreatedAt), string(o.Status), o.TotalTTC, strings.Join(itemsStr, "\n"))

	if o.CancelReason != "" {
		content += fmt.Sprintf("\n\nRaison de l'annulation: %s", o.CancelReason)
	}
	fmt.Fprintln(c.Out, orderStyle.Render(content))
}
