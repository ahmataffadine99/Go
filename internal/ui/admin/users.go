package adminui

import (
	"net/http"

	"ecommerce-cli/internal/models"
)

func (c *Client) fetchUsers() ([]models.User, error) {
	var users []models.User
	err := c.call("Chargement des utilisateurs...", http.MethodGet, "/api/admin/users", nil, &users)
	return users, err
}
