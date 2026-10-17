package components

import (
	"io"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type taskDoneMsg struct {
	err error
}

type spinnerModel struct {
	spinner spinner.Model
	label   string
	task    func() error
	err     error
	done    bool
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		return taskDoneMsg{err: m.task()}
	})
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case taskDoneMsg:
		m.done = true
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + m.label + "\n"
}

func RunWithSpinner(in io.Reader, out io.Writer, label string, task func() error) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	final, err := tea.NewProgram(spinnerModel{spinner: s, label: label, task: task}, tea.WithInput(in), tea.WithOutput(out)).Run()
	if err != nil {
		return task()
	}
	return final.(spinnerModel).err
}
