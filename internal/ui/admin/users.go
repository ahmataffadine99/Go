package adminui

import (
	"fmt"
	"net/http"
	"strings"

	"ecommerce-cli/internal/models"
	"ecommerce-cli/internal/ui/components"

	"github.com/charmbracelet/huh"
)

func (c *Client) fetchUsers() ([]models.User, error) {
	var users []models.User
	err := c.call("Chargement des utilisateurs...", http.MethodGet, "/api/admin/users", nil, &users)
	return users, err
}

func (c *Client) RunUsersMenu() {
	for {
		var action string
		err := c.runForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Gestion des utilisateurs").
				Options(
					huh.NewOption("Lister et gérer les utilisateurs", "list"),
					huh.NewOption("Créer un utilisateur", "create"),
					huh.NewOption("Retour", "back"),
				).
				Value(&action),
		))
		if err != nil || action == "back" {
			return
		}

		switch action {
		case "list":
			c.listUsers()
		case "create":
			c.createUser()
		}
	}
}

func confirmedLabel(confirmed bool) string {
	if confirmed {
		return "oui"
	}
	return "non"
}

func (c *Client) listUsers() {
	users, err := c.fetchUsers()
	if err != nil {
		c.printError(err)
		return
	}
	if len(users) == 0 {
		fmt.Fprintln(c.Out, "Aucun utilisateur.")
		return
	}

	rows := make([][]string, len(users))
	for i, u := range users {
		rows[i] = []string{fmt.Sprintf("%d", u.ID), u.Email, u.Role, confirmedLabel(u.IsConfirmed), formatDate(u.CreatedAt)}
	}

	idx, ok := components.RunTable(c.In, c.Out, "Utilisateurs", []string{"ID", "Email", "Rôle", "Confirmé", "Inscription"}, rows)
	if !ok {
		return
	}
	c.manageUser(users[idx])
}

func (c *Client) manageUser(u models.User) {
	options := []huh.Option[string]{huh.NewOption("Modifier", "edit")}
	if !u.IsConfirmed {
		options = append(options, huh.NewOption("Confirmer le compte", "confirm"))
	}
	options = append(options, huh.NewOption("Supprimer", "delete"), huh.NewOption("Retour", "back"))

	var action string
	if err := c.runForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Utilisateur " + u.Email).Options(options...).Value(&action),
	)); err != nil {
		return
	}

	switch action {
	case "edit":
		c.editUser(u)
	case "confirm":
		path := fmt.Sprintf("/api/admin/users/confirm?id=%d", u.ID)
		if err := c.call("Confirmation du compte...", http.MethodPost, path, nil, nil); err != nil {
			c.printError(err)
			return
		}
		c.printSuccess("Compte " + u.Email + " confirmé.")
	case "delete":
		if !c.confirm(fmt.Sprintf("Supprimer définitivement %s ?", u.Email)) {
			return
		}
		path := fmt.Sprintf("/api/admin/users/delete?id=%d", u.ID)
		if err := c.call("Suppression...", http.MethodDelete, path, nil, nil); err != nil {
			c.printError(err)
			return
		}
		c.printSuccess("Utilisateur " + u.Email + " supprimé.")
	}
}

func roleSelect(role *string) *huh.Select[string] {
	return huh.NewSelect[string]().
		Title("Rôle").
		Options(huh.NewOption("Client", "client"), huh.NewOption("Administrateur", "admin")).
		Value(role)
}

func (c *Client) editUser(u models.User) {
	email := u.Email
	role := u.Role
	var password string

	fmt.Fprintln(c.Out, titleStyle.Render("Modification de "+u.Email))
	if err := c.runForm(huh.NewGroup(
		huh.NewInput().Title("Email").Value(&email).Validate(huh.ValidateNotEmpty()),
		roleSelect(&role),
		huh.NewInput().Title("Nouveau mot de passe (laisser vide pour ne pas changer)").EchoMode(huh.EchoModePassword).Value(&password),
	)); err != nil {
		return
	}

	payload := models.AdminUserRequest{Email: strings.TrimSpace(email), Role: role, Password: password}
	path := fmt.Sprintf("/api/admin/users/update?id=%d", u.ID)
	if err := c.call("Enregistrement...", http.MethodPut, path, payload, nil); err != nil {
		c.printError(err)
		return
	}
	c.printSuccess("Utilisateur mis à jour.")
}

func (c *Client) createUser() {
	var email, password string
	role := "client"

	fmt.Fprintln(c.Out, titleStyle.Render("Nouvel utilisateur"))
	if err := c.runForm(huh.NewGroup(
		huh.NewInput().Title("Email").Value(&email).Validate(huh.ValidateNotEmpty()),
		huh.NewInput().Title("Mot de passe").EchoMode(huh.EchoModePassword).Value(&password).Validate(huh.ValidateNotEmpty()),
		roleSelect(&role),
	)); err != nil {
		return
	}

	var user models.User
	payload := models.AdminUserRequest{Email: strings.TrimSpace(email), Password: password, Role: role}
	if err := c.call("Création de l'utilisateur...", http.MethodPost, "/api/admin/users/create", payload, &user); err != nil {
		c.printError(err)
		return
	}
	c.printSuccess(fmt.Sprintf("Utilisateur %s créé (ID %d, compte confirmé).", user.Email, user.ID))
}
