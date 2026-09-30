package productsui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"ecommerce-cli/internal/models"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type ProductClient struct {
	BaseURL string
}

func NewProductClient(baseURL string) *ProductClient {
	return &ProductClient{BaseURL: baseURL}
}

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	cardStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(1).Margin(1)
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	priceStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	codeStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func (c *ProductClient) RunProductSearchMenu() error {
	var query string
	var category string
	var minPriceStr string
	var maxPriceStr string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Recherche (nom, description, identifiant métier)").
				Placeholder("ex: PC, PDT-LAP001...").
				Value(&query),
			huh.NewInput().
				Title("Catégorie (optionnel)").
				Placeholder("ex: Informatique, Accessoires...").
				Value(&category),
			huh.NewInput().
				Title("Prix Minimum (€)").
				Placeholder("0").
				Value(&minPriceStr),
			huh.NewInput().
				Title("Prix Maximum (€)").
				Placeholder("1000").
				Value(&maxPriceStr),
		),
	)

	fmt.Println(headerStyle.Render("Recherche de Produits"))
	if err := form.Run(); err != nil {
		return err
	}

	params := url.Values{}
	if query != "" {
		params.Add("q", query)
	}
	if category != "" {
		params.Add("category", category)
	}
	if minPriceStr != "" {
		params.Add("min_price", minPriceStr)
	}
	if maxPriceStr != "" {
		params.Add("max_price", maxPriceStr)
	}

	reqURL := fmt.Sprintf("%s/api/products?%s", c.BaseURL, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		fmt.Println("Erreur lors de la récupération des produits.")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Erreur serveur:", string(body))
		return fmt.Errorf("search failed")
	}

	var products []models.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		fmt.Println("Erreur de lecture des données.")
		return err
	}

	if len(products) == 0 {
		fmt.Println("\nAucun produit trouvé correspondant à vos critères.")
		return nil
	}

	fmt.Printf("\n--- %d Produit(s) trouvé(s) ---\n\n", len(products))

	for _, p := range products {
		priceTTC := p.Price
		cardContent := fmt.Sprintf(
			"%s [%s]\n%s\nCatégorie: %s | Stock: %d\nPrix TTC: %s",
			titleStyle.Render(p.Name),
			codeStyle.Render(p.BusinessID),
			p.Description,
			p.Category,
			p.Stock,
			priceStyle.Render(fmt.Sprintf("%.2f €", priceTTC)),
		)
		fmt.Println(cardStyle.Render(cardContent))
	}

	return nil
}

func (c *ProductClient) ListAllProducts() ([]models.Product, error) {
	resp, err := http.Get(c.BaseURL + "/api/products")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var products []models.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		return nil, err
	}
	return products, nil
}

func parseOptionalFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}
