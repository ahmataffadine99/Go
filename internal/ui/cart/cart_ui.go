package cartui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ecommerce-cli/internal/models"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type CartClient struct {
	BaseURL string
	Token   string
}

func NewCartClient(baseURL string) *CartClient {
	return &CartClient{BaseURL: baseURL}
}

var (
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).MarginBottom(1)
	itemStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("63")).Padding(1).Margin(1)
	totalStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")).MarginTop(1)
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("46"))
)

func (c *CartClient) doAuthRequest(method, endpoint string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, c.BaseURL+endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{}
	return client.Do(req)
}

func (c *CartClient) RunCartMenu() error {
	for {
		cart, err := c.fetchCart()
		if err != nil {
			fmt.Println(errorStyle.Render("Erreur de récupération du panier: " + err.Error()))
			return err
		}

		c.displayCart(cart)

		var action string
		options := []huh.Option[string]{
			huh.NewOption("Ajouter un produit", "add"),
		}

		if len(cart.Items) > 0 {
			options = append(options,
				huh.NewOption("Modifier la quantité d'un produit", "update"),
				huh.NewOption("Retirer un produit", "remove"),
				huh.NewOption("Passer à la caisse (Payer)", "pay"),
			)
		}
		options = append(options, huh.NewOption("Retour au menu principal", "back"))

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Action Panier").
					Options(options...).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			return nil
		}

		switch action {
		case "add":
			c.runAddItem()
		case "update":
			c.runUpdateItem()
		case "remove":
			c.runRemoveItem()
		case "pay":
			success := c.runPay(cart)
			if success {
				return nil // Go back to main menu after payment
			}
		case "back":
			return nil
		}
	}
}

func (c *CartClient) fetchCart() (*models.Cart, error) {
	resp, err := c.doAuthRequest(http.MethodGet, "/api/cart", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get cart, status: %d", resp.StatusCode)
	}

	var cart models.Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err != nil {
		return nil, err
	}
	return &cart, nil
}

func (c *CartClient) displayCart(cart *models.Cart) {
	fmt.Println(headerStyle.Render("\n=== Votre Panier ==="))
	if len(cart.Items) == 0 {
		fmt.Println("Votre panier est actuellement vide.")
		return
	}

	for _, item := range cart.Items {
		content := fmt.Sprintf("Produit ID: %d | Quantité: %d", item.ProductID, item.Quantity)
		fmt.Println(itemStyle.Render(content))
	}
	fmt.Println(totalStyle.Render(fmt.Sprintf("Total TTC: %.2f €", cart.TotalTTC)))
}

func (c *CartClient) runAddItem() {
	var productIDStr string
	var quantityStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ID du produit (numéro)").
				Value(&productIDStr),
			huh.NewInput().
				Title("Quantité").
				Value(&quantityStr),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	var productID, quantity int
	fmt.Sscanf(productIDStr, "%d", &productID)
	fmt.Sscanf(quantityStr, "%d", &quantity)

	payload, _ := json.Marshal(map[string]int{
		"product_id": productID,
		"quantity":   quantity,
	})

	resp, err := c.doAuthRequest(http.MethodPost, "/api/cart/items", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Erreur: " + string(body)))
		return
	}

	fmt.Println(successStyle.Render("Produit ajouté avec succès !"))
}

func (c *CartClient) runUpdateItem() {
	var productIDStr string
	var quantityStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ID du produit à modifier").
				Value(&productIDStr),
			huh.NewInput().
				Title("Nouvelle quantité").
				Value(&quantityStr),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	var productID, quantity int
	fmt.Sscanf(productIDStr, "%d", &productID)
	fmt.Sscanf(quantityStr, "%d", &quantity)

	payload, _ := json.Marshal(map[string]int{
		"quantity": quantity,
	})

	endpoint := fmt.Sprintf("/api/cart/items?product_id=%d", productID)
	resp, err := c.doAuthRequest(http.MethodPut, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Erreur: " + string(body)))
		return
	}

	fmt.Println(successStyle.Render("Quantité modifiée !"))
}

func (c *CartClient) runRemoveItem() {
	var productIDStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("ID du produit à retirer").
				Value(&productIDStr),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	var productID int
	fmt.Sscanf(productIDStr, "%d", &productID)

	endpoint := fmt.Sprintf("/api/cart/items?product_id=%d", productID)
	resp, err := c.doAuthRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Erreur: " + string(body)))
		return
	}

	fmt.Println(successStyle.Render("Produit retiré du panier !"))
}

func (c *CartClient) runPay(cart *models.Cart) bool {
	var card, cvc, exp string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Numéro de carte (16 chiffres)").Value(&card),
			huh.NewInput().Title("Date d'expiration (MM/YY)").Value(&exp),
			huh.NewInput().Title("CVC (3 chiffres)").Value(&cvc),
		),
	)

	if err := form.Run(); err != nil {
		return false
	}

	payload, _ := json.Marshal(map[string]string{
		"card_number": card,
		"expiry_date": exp,
		"cvc":         cvc,
	})

	resp, err := c.doAuthRequest(http.MethodPost, "/api/cart/pay", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Paiement refusé: " + string(body)))
		return false
	}

	fmt.Println(successStyle.Render("Paiement accepté ! Commande validée."))
	return true
}
