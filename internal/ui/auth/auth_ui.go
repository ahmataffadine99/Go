package authui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type AuthClient struct {
	BaseURL string
	Token   string
	Email   string
	Role    string
}

func NewAuthClient(baseURL string) *AuthClient {
	return &AuthClient{BaseURL: baseURL}
}

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

func (c *AuthClient) RunLoginForm() error {
	var email, password string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Adresse Email").
				Placeholder("user@example.com").
				Value(&email),
			huh.NewInput().
				Title("Mot de passe").
				EchoMode(huh.EchoModePassword).
				Value(&password),
		),
	)

	fmt.Println(titleStyle.Render("Connexion Utilisateur"))
	if err := form.Run(); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(c.BaseURL+"/api/auth/login", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur de connexion au serveur Backend HTTP"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Echec de la connexion: " + string(body)))
		return fmt.Errorf("login failed")
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if token, ok := res["token"].(string); ok {
		c.Token = token
		c.Email = email
		if userMap, ok := res["user"].(map[string]interface{}); ok {
			if role, ok := userMap["role"].(string); ok {
				c.Role = role
			}
		}
		fmt.Println(successStyle.Render("Connexion réussie ! Token enregistré."))
	}

	return nil
}

func (c *AuthClient) RunRegisterForm() error {
	var email, password string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email pour inscription").
				Placeholder("user@example.com").
				Value(&email),
			huh.NewInput().
				Title("Mot de passe").
				EchoMode(huh.EchoModePassword).
				Value(&password),
		),
	)

	fmt.Println(titleStyle.Render("Inscription"))
	if err := form.Run(); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})

	resp, err := http.Post(c.BaseURL+"/api/auth/register", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur de connexion au serveur Backend HTTP"))
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Println(errorStyle.Render("Erreur lors de l'inscription: " + string(body)))
		return fmt.Errorf("registration failed")
	}

	var res map[string]interface{}
	json.Unmarshal(body, &res)

	code := ""
	if codeVal, ok := res["confirmation_code"].(string); ok {
		code = codeVal
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("Inscription réussie ! Votre code de confirmation est : %s", code)))
	return c.RunConfirmFormWithEmail(email, code)
}

func (c *AuthClient) RunConfirmFormWithEmail(defaultEmail, defaultCode string) error {
	email := defaultEmail
	code := defaultCode

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email").
				Value(&email),
			huh.NewInput().
				Title("Code de confirmation").
				Value(&code),
		),
	)

	fmt.Println(titleStyle.Render("Confirmation de compte"))
	if err := form.Run(); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"email": email,
		"code":  code,
	})

	resp, err := http.Post(c.BaseURL+"/api/auth/confirm", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Confirmation échouée: " + string(body)))
		return fmt.Errorf("confirmation failed")
	}

	fmt.Println(successStyle.Render("Compte confirmé avec succès ! Vous pouvez maintenant vous connecter."))
	return nil
}

func (c *AuthClient) RunResetPasswordForm() error {
	var email, newPassword string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Email du compte à réinitialiser").
				Value(&email),
			huh.NewInput().
				Title("Nouveau mot de passe").
				EchoMode(huh.EchoModePassword).
				Value(&newPassword),
		),
	)

	fmt.Println(titleStyle.Render("Réinitialisation du mot de passe"))
	if err := form.Run(); err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]string{
		"email":        email,
		"new_password": newPassword,
	})

	resp, err := http.Post(c.BaseURL+"/api/auth/reset-password", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(errorStyle.Render("Erreur réseau"))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(errorStyle.Render("Echec réinitialisation: " + string(body)))
		return fmt.Errorf("reset failed")
	}

	fmt.Println(successStyle.Render("Mot de passe réinitialisé avec succès !"))
	return nil
}
