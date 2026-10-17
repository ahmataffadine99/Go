package adminui

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/ui/components"

	"github.com/charmbracelet/huh"
)

func (c *Client) fetchProducts() ([]models.Product, error) {
	var products []models.Product
	err := c.call("Chargement des produits...", http.MethodGet, "/api/products", nil, &products)
	return products, err
}

func (c *Client) RunProductsMenu() {
	for {
		var action string
		err := c.runForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Gestion des produits").
				Options(
					huh.NewOption("Lister les produits", "list"),
					huh.NewOption("Ajouter un produit", "add"),
					huh.NewOption("Retour", "back"),
				).
				Value(&action),
		))
		if err != nil || action == "back" {
			return
		}

		switch action {
		case "list":
			c.listProducts()
		case "add":
			c.addProduct()
		}
	}
}

func (c *Client) listProducts() {
	products, err := c.fetchProducts()
	if err != nil {
		c.printError(err)
		return
	}
	if len(products) == 0 {
		fmt.Fprintln(c.Out, "Aucun produit dans le catalogue.")
		return
	}

	rows := make([][]string, len(products))
	for i, p := range products {
		rows[i] = []string{
			p.BusinessID,
			p.Name,
			p.Category,
			fmt.Sprintf("%.2f €", p.Price),
			fmt.Sprintf("%.2f €", p.PriceTTC),
			strconv.Itoa(p.Stock),
		}
	}

	idx, ok := components.RunTable(c.In, c.Out, "Catalogue produits", []string{"Identifiant", "Nom", "Catégorie", "Prix HT", "Prix TTC", "Stock"}, rows)
	if ok {
		p := products[idx]
		fmt.Fprintln(c.Out, cardStyle.Render(fmt.Sprintf("%s - %s\n%s\nCatégorie: %s\nPrix HT: %.2f € | Prix TTC: %.2f €\nStock: %d",
			p.BusinessID, p.Name, p.Description, p.Category, p.Price, p.PriceTTC, p.Stock)))
	}
}

func validateNumber(isInt bool) func(string) error {
	return func(s string) error {
		s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
		if isInt {
			if v, err := strconv.Atoi(s); err != nil || v < 0 {
				return fmt.Errorf("entier positif attendu")
			}
			return nil
		}
		if v, err := strconv.ParseFloat(s, 64); err != nil || v <= 0 {
			return fmt.Errorf("nombre positif attendu")
		}
		return nil
	}
}

func (c *Client) addProduct() {
	var name, description, category, priceStr, stockStr string

	fmt.Fprintln(c.Out, titleStyle.Render("Nouveau produit"))
	err := c.runForm(huh.NewGroup(
		huh.NewInput().Title("Nom").Value(&name).Validate(huh.ValidateNotEmpty()),
		huh.NewText().Title("Description").Value(&description),
		huh.NewInput().Title("Catégorie").Value(&category).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Prix HT (€)").Value(&priceStr).Validate(validateNumber(false)),
		huh.NewInput().Title("Stock").Value(&stockStr).Validate(validateNumber(true)),
	))
	if err != nil {
		return
	}

	price, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(priceStr), ",", "."), 64)
	stock, _ := strconv.Atoi(strings.TrimSpace(stockStr))

	payload := models.CreateProductRequest{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Price:       price,
		Category:    strings.TrimSpace(category),
		Stock:       stock,
	}

	var product models.Product
	if err := c.call("Création du produit...", http.MethodPost, "/api/admin/products", payload, &product); err != nil {
		c.printError(err)
		return
	}

	c.printSuccess(fmt.Sprintf("Produit créé: %s (%s) - %.2f € TTC", product.Name, product.BusinessID, product.PriceTTC))
}
