package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type adminModel struct {
	choices []string
	cursor  int
}

func initialAdminModel() adminModel {
	return adminModel{
		choices: []string{"Gérer les Commandes", "Gérer les Utilisateurs", "Ajouter un Produit", "Quitter"},
	}
}

func (m adminModel) Init() tea.Cmd {
	return nil
}

func (m adminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.choices)-1 {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m adminModel) View() string {
	s := "--- E-Commerce Admin CLI ---\n\n"

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nAppuyez sur q pour quitter.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialAdminModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running admin application: %v\n", err)
		os.Exit(1)
	}
}
