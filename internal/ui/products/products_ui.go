package productsui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"ecommerce-cli/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type ProductClient struct {
	BaseURL string
	In      io.Reader
	Out     io.Writer
}

func NewProductClient(baseURL string, in io.Reader, out io.Writer) *ProductClient {
	return &ProductClient{BaseURL: baseURL, In: in, Out: out}
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
	priceType := "ttc"

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
			huh.NewSelect[string]().
				Title("Filtrer le prix en").
				Options(
					huh.NewOption("Prix TTC", "ttc"),
					huh.NewOption("Prix HT", "ht"),
				).
				Value(&priceType),
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

	fmt.Fprintln(c.Out, headerStyle.Render("Recherche de Produits"))
	if err := form.WithProgramOptions(tea.WithInput(c.In), tea.WithOutput(c.Out)).Run(); err != nil {
		return err
	}

	params := url.Values{}
	if query != "" {
		params.Add("q", query)
	}
	if category != "" {
		params.Add("category", category)
	}
	suffix := ""
	if priceType == "ttc" {
		suffix = "_ttc"
	}
	if minPriceStr != "" {
		params.Add("min_price"+suffix, minPriceStr)
	}
	if maxPriceStr != "" {
		params.Add("max_price"+suffix, maxPriceStr)
	}

	reqURL := fmt.Sprintf("%s/api/products?%s", c.BaseURL, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		fmt.Fprintln(c.Out, "Erreur lors de la récupération des produits.")
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintln(c.Out, "Erreur serveur:", string(body))
		return fmt.Errorf("search failed")
	}

	var products []models.Product
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		fmt.Fprintln(c.Out, "Erreur de lecture des données.")
		return err
	}

	if len(products) == 0 {
		fmt.Fprintln(c.Out, "\nAucun produit trouvé correspondant à vos critères.")
		return nil
	}

	fmt.Fprintf(c.Out, "\n--- %d Produit(s) trouvé(s) ---\n\n", len(products))

	for _, p := range products {
		cardContent := fmt.Sprintf(
			"ID: %d | %s [%s]\n%s\nCatégorie: %s | Stock: %d\nPrix HT: %.2f € | Prix TTC: %s",
			p.ID,
			titleStyle.Render(p.Name),
			codeStyle.Render(p.BusinessID),
			p.Description,
			p.Category,
			p.Stock,
			p.Price,
			priceStyle.Render(fmt.Sprintf("%.2f €", p.PriceTTC)),
		)
		fmt.Fprintln(c.Out, cardStyle.Render(cardContent))
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
