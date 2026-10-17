package components

import (
	"io"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	tableTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	tableBorder     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("63"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

const maxColumnWidth = 40

type tableModel struct {
	title    string
	table    table.Model
	selected int
	chosen   bool
	quitting bool
}

func (m tableModel) Init() tea.Cmd {
	return nil
}

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			if len(m.table.Rows()) > 0 {
				m.chosen = true
				m.selected = m.table.Cursor()
			}
			m.quitting = true
			return m, tea.Quit
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() string {
	if m.quitting {
		return ""
	}
	return tableTitleStyle.Render(m.title) + "\n" +
		tableBorder.Render(m.table.View()) + "\n" +
		helpStyle.Render("haut/bas: naviguer | entrée: sélectionner | q: retour") + "\n"
}

func RunTable(in io.Reader, out io.Writer, title string, headers []string, rows [][]string) (int, bool) {
	columns := make([]table.Column, len(headers))
	for i, h := range headers {
		width := lipgloss.Width(h)
		for _, row := range rows {
			if i < len(row) && lipgloss.Width(row[i]) > width {
				width = lipgloss.Width(row[i])
			}
		}
		columns[i] = table.Column{Title: h, Width: min(width+1, maxColumnWidth)}
	}

	tableRows := make([]table.Row, len(rows))
	for i, row := range rows {
		tableRows[i] = table.Row(row)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(tableRows),
		table.WithFocused(true),
		table.WithHeight(min(len(rows), 12)+1),
	)

	styles := table.DefaultStyles()
	styles.Header = styles.Header.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).Bold(true)
	styles.Selected = styles.Selected.Foreground(lipgloss.Color("229")).Background(lipgloss.Color("57"))
	t.SetStyles(styles)

	final, err := tea.NewProgram(tableModel{title: title, table: t}, tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return 0, false
	}
	m := final.(tableModel)
	return m.selected, m.chosen
}
