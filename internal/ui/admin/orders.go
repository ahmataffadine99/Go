package adminui

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/ui/components"

	"github.com/charmbracelet/huh"
)

var orderStatuses = []models.OrderStatus{
	models.OrderStatusPending,
	models.OrderStatusPaid,
	models.OrderStatusShipping,
	models.OrderStatusDelivered,
	models.OrderStatusCanceled,
}

func formatDate(s string) string {
	s = strings.Replace(s, "T", " ", 1)
	if len(s) > 16 {
		return s[:16]
	}
	return s
}

func (c *Client) usersByID() map[int64]string {
	users, err := c.fetchUsers()
	result := make(map[int64]string)
	if err != nil {
		return result
	}
	for _, u := range users {
		result[u.ID] = u.Email
	}
	return result
}

func (c *Client) RunOrdersMenu() {
	for {
		var action string
		err := c.runForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Gestion des commandes").
				Options(
					huh.NewOption("Lister et modifier les commandes", "list"),
					huh.NewOption("Créer une commande pour un client", "create"),
					huh.NewOption("Retour", "back"),
				).
				Value(&action),
		))
		if err != nil || action == "back" {
			return
		}

		switch action {
		case "list":
			c.listOrders()
		case "create":
			c.createOrder()
		}
	}
}

func (c *Client) listOrders() {
	var orders []models.Order
	if err := c.call("Chargement des commandes...", http.MethodGet, "/api/admin/orders", nil, &orders); err != nil {
		c.printError(err)
		return
	}
	if len(orders) == 0 {
		fmt.Fprintln(c.Out, "Aucune commande enregistrée.")
		return
	}

	emails := c.usersByID()
	rows := make([][]string, len(orders))
	for i, o := range orders {
		client := emails[o.UserID]
		if client == "" {
			client = "#" + strconv.FormatInt(o.UserID, 10)
		}
		rows[i] = []string{o.BusinessID, client, formatDate(o.CreatedAt), fmt.Sprintf("%.2f €", o.TotalTTC), string(o.Status)}
	}

	idx, ok := components.RunTable(c.In, c.Out, "Commandes", []string{"Identifiant", "Client", "Date", "Total TTC", "Statut"}, rows)
	if !ok {
		return
	}

	order := orders[idx]
	c.printOrder(order, emails[order.UserID])

	if c.confirm("Changer le statut de cette commande ?") {
		c.updateOrderStatus(order)
	}
}

func (c *Client) printOrder(o models.Order, email string) {
	var items []string
	for _, item := range o.Items {
		name := "Produit inconnu"
		if item.Product != nil {
			name = item.Product.Name
		}
		items = append(items, fmt.Sprintf("  - %s (x%d) : %.2f € TTC l'unité", name, item.Quantity, item.UnitPrice))
	}

	content := fmt.Sprintf("Commande %s\nClient: %s\nDate: %s\nStatut: %s\nTotal TTC: %.2f €\n\nProduits:\n%s",
		o.BusinessID, email, formatDate(o.CreatedAt), o.Status, o.TotalTTC, strings.Join(items, "\n"))
	if o.CancelReason != "" {
		content += "\n\nRaison de l'annulation: " + o.CancelReason
	}
	fmt.Fprintln(c.Out, cardStyle.Render(content))
}

func (c *Client) updateOrderStatus(o models.Order) {
	status := o.Status
	options := make([]huh.Option[models.OrderStatus], len(orderStatuses))
	for i, s := range orderStatuses {
		options[i] = huh.NewOption(string(s), s)
	}

	if err := c.runForm(huh.NewGroup(
		huh.NewSelect[models.OrderStatus]().Title("Nouveau statut").Options(options...).Value(&status),
	)); err != nil {
		return
	}

	reason := ""
	if status == models.OrderStatusCanceled {
		reason = o.CancelReason
		if err := c.runForm(huh.NewGroup(
			huh.NewInput().Title("Raison de l'annulation").Value(&reason).Validate(huh.ValidateNotEmpty()),
		)); err != nil {
			return
		}
	}

	payload := models.UpdateOrderStatusRequest{Status: status, CancelReason: strings.TrimSpace(reason)}
	path := fmt.Sprintf("/api/admin/orders/status?id=%d", o.ID)
	if err := c.call("Mise à jour du statut...", http.MethodPut, path, payload, nil); err != nil {
		c.printError(err)
		return
	}
	c.printSuccess(fmt.Sprintf("Commande %s passée au statut \"%s\".", o.BusinessID, status))
}

func (c *Client) createOrder() {
	users, err := c.fetchUsers()
	if err != nil {
		c.printError(err)
		return
	}
	if len(users) == 0 {
		c.printError(errors.New("aucun utilisateur disponible"))
		return
	}

	products, err := c.fetchProducts()
	if err != nil {
		c.printError(err)
		return
	}
	if len(products) == 0 {
		c.printError(errors.New("aucun produit disponible"))
		return
	}

	userOptions := make([]huh.Option[int64], len(users))
	for i, u := range users {
		userOptions[i] = huh.NewOption(fmt.Sprintf("%s (%s)", u.Email, u.Role), u.ID)
	}

	var userID int64
	fmt.Fprintln(c.Out, titleStyle.Render("Nouvelle commande"))
	if err := c.runForm(huh.NewGroup(
		huh.NewSelect[int64]().Title("Client").Options(userOptions...).Value(&userID),
	)); err != nil {
		return
	}

	productByID := make(map[int64]models.Product)
	productOptions := make([]huh.Option[int64], 0, len(products)+1)
	for _, p := range products {
		productByID[p.ID] = p
		productOptions = append(productOptions, huh.NewOption(fmt.Sprintf("%s - %s (%.2f € TTC) | Stock: %d", p.BusinessID, p.Name, p.PriceTTC, p.Stock), p.ID))
	}
	productOptions = append(productOptions, huh.NewOption("Terminer la sélection", int64(0)))

	var items []models.AddToCartRequest
	for {
		var productID int64
		var quantityStr string
		if err := c.runForm(huh.NewGroup(
			huh.NewSelect[int64]().Title(fmt.Sprintf("Produit à ajouter (%d ligne(s))", len(items))).Options(productOptions...).Value(&productID),
		)); err != nil {
			return
		}
		if productID == 0 {
			break
		}

		quantityStr = "1"
		if err := c.runForm(huh.NewGroup(
			huh.NewInput().Title("Quantité de " + productByID[productID].Name).Value(&quantityStr).Validate(func(s string) error {
				if q, err := strconv.Atoi(strings.TrimSpace(s)); err != nil || q <= 0 {
					return errors.New("entier strictement positif attendu")
				}
				return nil
			}),
		)); err != nil {
			return
		}

		quantity, _ := strconv.Atoi(strings.TrimSpace(quantityStr))
		items = append(items, models.AddToCartRequest{ProductID: productID, Quantity: quantity})
	}

	if len(items) == 0 {
		fmt.Fprintln(c.Out, "Commande annulée: aucun produit sélectionné.")
		return
	}

	var total float64
	var lines []string
	for _, item := range items {
		p := productByID[item.ProductID]
		total += p.PriceTTC * float64(item.Quantity)
		lines = append(lines, fmt.Sprintf("  - %s (x%d)", p.Name, item.Quantity))
	}
	fmt.Fprintln(c.Out, cardStyle.Render(fmt.Sprintf("Récapitulatif\n%s\nTotal TTC estimé: %.2f €", strings.Join(lines, "\n"), total)))

	if !c.confirm("Valider la création de la commande ?") {
		return
	}

	var order models.Order
	payload := models.AdminCreateOrderRequest{UserID: userID, Items: items}
	if err := c.call("Création de la commande...", http.MethodPost, "/api/admin/orders/create", payload, &order); err != nil {
		c.printError(err)
		return
	}
	c.printSuccess(fmt.Sprintf("Commande %s créée (statut: %s, total: %.2f € TTC).", order.BusinessID, order.Status, order.TotalTTC))
}
