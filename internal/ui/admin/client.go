package adminui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ecommerce-cli/internal/ui/components"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	successStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	cardStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63")).Padding(0, 1)
)

type Client struct {
	BaseURL    string
	Token      string
	Email      string
	In         io.Reader
	Out        io.Writer
	httpClient *http.Client
}

func NewClient(baseURL string, in io.Reader, out io.Writer) *Client {
	return &Client{
		BaseURL:    baseURL,
		In:         in,
		Out:        out,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) request(method, path string, payload, result interface{}) error {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("serveur injoignable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		var apiErr struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error != "" {
			return errors.New(apiErr.Error)
		}
		return fmt.Errorf("erreur serveur (%d)", resp.StatusCode)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) call(label, method, path string, payload, result interface{}) error {
	return components.RunWithSpinner(c.In, c.Out, label, func() error {
		return c.request(method, path, payload, result)
	})
}

func (c *Client) runForm(groups ...*huh.Group) error {
	return huh.NewForm(groups...).WithProgramOptions(tea.WithInput(c.In), tea.WithOutput(c.Out)).Run()
}

func (c *Client) confirm(title string) bool {
	var ok bool
	if err := c.runForm(huh.NewGroup(huh.NewConfirm().Title(title).Affirmative("Oui").Negative("Non").Value(&ok))); err != nil {
		return false
	}
	return ok
}

func (c *Client) printSuccess(msg string) {
	fmt.Fprintln(c.Out, successStyle.Render(msg))
}

func (c *Client) printError(err error) {
	fmt.Fprintln(c.Out, errorStyle.Render("Erreur: "+err.Error()))
}

func (c *Client) Login() error {
	var email, password string

	fmt.Fprintln(c.Out, titleStyle.Render("Connexion Administrateur"))
	err := c.runForm(huh.NewGroup(
		huh.NewInput().Title("Adresse Email").Placeholder("admin@shop.local").Value(&email),
		huh.NewInput().Title("Mot de passe").EchoMode(huh.EchoModePassword).Value(&password),
	))
	if err != nil {
		return err
	}

	var res struct {
		Token string `json:"token"`
		User  struct {
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"user"`
	}
	payload := map[string]string{"email": strings.TrimSpace(email), "password": password}
	if err := c.call("Connexion en cours...", http.MethodPost, "/api/auth/login", payload, &res); err != nil {
		c.printError(err)
		return err
	}

	if res.User.Role != "admin" {
		err := errors.New("ce compte n'a pas les droits administrateur")
		c.printError(err)
		return err
	}

	c.Token = res.Token
	c.Email = res.User.Email
	c.printSuccess("Connexion réussie.")
	return nil
}

func (c *Client) Logout() {
	c.Token = ""
	c.Email = ""
}
