package app

import (
	"fmt"
	"io"

	adminui "ecommerce-cli/internal/ui/admin"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func RunAdminApp(baseURL string, in io.Reader, out io.Writer) {
	client := adminui.NewClient(baseURL, in, out)

	for {
		menuTitle := "--- E-Commerce Admin CLI ---"
		if client.Email != "" {
			menuTitle = fmt.Sprintf("--- E-Commerce Admin CLI (%s) ---", client.Email)
		}
		fmt.Fprintln(out, menuTitle)

		var options []huh.Option[string]
		if client.Token == "" {
			options = []huh.Option[string]{
				huh.NewOption("Connexion", "login"),
				huh.NewOption("Quitter", "quit"),
			}
		} else {
			options = []huh.Option[string]{
				huh.NewOption("Gérer les produits", "products"),
				huh.NewOption("Déconnexion", "logout"),
				huh.NewOption("Quitter", "quit"),
			}
		}

		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Menu administrateur").
					Options(options...).
					Value(&action),
			),
		).WithProgramOptions(tea.WithInput(in), tea.WithOutput(out))

		if err := form.Run(); err != nil {
			fmt.Fprintln(out, "Au revoir.")
			return
		}

		switch action {
		case "login":
			_ = client.Login()
		case "products":
			client.RunProductsMenu()
		case "logout":
			client.Logout()
			fmt.Fprintln(out, "Déconnexion réussie.")
		case "quit":
			fmt.Fprintln(out, "Au revoir.")
			return
		}

		fmt.Fprintln(out, "")
	}
}
