package app

import (
	"fmt"
	"io"
	"os"

	authui "ecommerce-cli/internal/ui/auth"
	cartui "ecommerce-cli/internal/ui/cart"
	ordersui "ecommerce-cli/internal/ui/orders"
	productsui "ecommerce-cli/internal/ui/products"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func RunClientApp(baseURL string, in io.Reader, out io.Writer) {
	authClient := authui.NewAuthClient(baseURL, in, out)
	productClient := productsui.NewProductClient(baseURL, in, out)
	cartClient := cartui.NewCartClient(baseURL, in, out)
	orderClient := ordersui.NewOrderClient(baseURL, "", in, out)

	for {
		var action string

		menuTitle := "--- E-Commerce Client CLI ---"
		if authClient.Email != "" {
			menuTitle = fmt.Sprintf("--- E-Commerce Client CLI (%s) ---", authClient.Email)
		}

		fmt.Fprintln(out, menuTitle)

		var options []huh.Option[string]
		if authClient.Token == "" {
			options = []huh.Option[string]{
				huh.NewOption("Parcourir / Rechercher des produits", "products"),
				huh.NewOption("Connexion", "login"),
				huh.NewOption("Inscription", "register"),
				huh.NewOption("Confirmation de compte", "confirm"),
				huh.NewOption("Mot de passe oublié", "reset"),
				huh.NewOption("Quitter", "quit"),
			}
		} else {
			options = []huh.Option[string]{
				huh.NewOption("Parcourir / Rechercher des produits", "products"),
				huh.NewOption("Mon Panier", "cart"),
				huh.NewOption("Mes Commandes", "orders"),
				huh.NewOption("Déconnexion", "logout"),
				huh.NewOption("Quitter", "quit"),
			}
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Menu principal").
					Options(options...).
					Value(&action),
			),
		).WithProgramOptions(tea.WithInput(in), tea.WithOutput(out))

		if err := form.Run(); err != nil {
			fmt.Fprintln(out, "Au revoir.")
			return
		}

		switch action {
		case "products":
			_ = productClient.RunProductSearchMenu()
		case "login":
			_ = authClient.RunLoginForm()
		case "register":
			_ = authClient.RunRegisterForm()
		case "confirm":
			_ = authClient.RunConfirmFormWithEmail("", "")
		case "reset":
			_ = authClient.RunResetPasswordForm()
		case "logout":
			authClient.Token = ""
			authClient.Email = ""
			authClient.Role = ""
			fmt.Fprintln(out, "Déconnexion réussie.")
		case "quit":
			fmt.Fprintln(out, "Au revoir.")
			// If running in terminal, we exit. If in SSH, we return to close session.
			if in == os.Stdin {
				os.Exit(0)
			}
			return
		case "cart":
			cartClient.Token = authClient.Token
			_ = cartClient.RunCartMenu()
		case "orders":
			orderClient.Token = authClient.Token
			orderClient.RunOrdersMenu()
		}

		fmt.Fprintln(out, "")
	}
}
