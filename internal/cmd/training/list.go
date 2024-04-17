package training

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/replicate/cli/internal/client"
	"github.com/replicate/cli/internal/util"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			selected := m.table.SelectedRow()
			if len(selected) == 0 {
				return m, nil
			}
			url := fmt.Sprintf("https://replicate.com/p/%s", selected[0])
			return m, tea.ExecProcess(exec.Command("open", url), nil)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List trainings",
	RunE: func(cmd *cobra.Command, _ []string) error {
		ctx := cmd.Context()

		r8, err := client.NewClient()
		if err != nil {
			return err
		}

		trainings, err := r8.ListTrainings(ctx)
		if err != nil {
			return fmt.Errorf("failed to get trainings: %w", err)
		}

		columns := []table.Column{
			{Title: "ID", Width: 20},
			{Title: "Version", Width: 20},
			{Title: "", Width: 3},
			{Title: "Created", Width: 20},
		}

		rows := []table.Row{}

		for _, training := range trainings.Results {
			rows = append(rows, table.Row{
				training.ID,
				training.Version,
				util.StatusSymbol(training.Status),
				training.CreatedAt,
			})
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(30),
		)

		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)
		t.SetStyles(s)

		m := model{t}
		if _, err := tea.NewProgram(m).Run(); err != nil {
			return err
		}

		return nil
	},
}
