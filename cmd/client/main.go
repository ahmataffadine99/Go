package main

import (
	"fmt"
	"os"

	authui "ecommerce-cli/internal/ui/auth"

	"github.com/charmbracelet/huh"
)

func main() {
	authClient := authui.NewAuthClient("http://localhost:8080")

	for {
		var action string

		menuTitle := "--- E-Commerce Client CLI ---"
		if authClient.Email != "" {
			menuTitle = fmt.Sprintf("--- E-Commerce Client CLI (%s) ---", authClient.Email)
		}

		fmt.Println(menuTitle)

		var options []huh.Option[string]
		if authClient.Token == "" {
			options = []huh.Option[string]{
				huh.NewOption("Connexion", "login"),
				huh.NewOption("Inscription", "register"),
				huh.NewOption("Confirmation de compte", "confirm"),
				huh.NewOption("Mot de passe oublié", "reset"),
				huh.NewOption("Quitter", "quit"),
			}
		} else {
			options = []huh.Option[string]{
				huh.NewOption("Parcourir les produits", "products"),
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
		)

		if err := form.Run(); err != nil {
			fmt.Println("Au revoir.")
			os.Exit(0)
		}

		switch action {
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
			fmt.Println("Déconnexion réussie.")
		case "quit":
			fmt.Println("Au revoir.")
			os.Exit(0)
		case "products":
			fmt.Println("Module Produits (en cours de développement)")
		case "cart":
			fmt.Println("Module Panier (en cours de développement)")
		case "orders":
			fmt.Println("Module Commandes (en cours de développement)")
		}

		fmt.Println()
	}
}
