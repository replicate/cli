package prediction

import (
	"fmt"
	"os/exec"

	"github.com/replicate/cli/internal/util"
	"github.com/replicate/replicate-go"
	"github.com/spf13/cobra"

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
	switch msg := msg.(type) {
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
			url := fmt.Sprintf("https://replicate.com/p/%s", m.table.SelectedRow()[0])
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
	Short: "List predictions",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := replicate.NewClient(replicate.WithTokenFromEnv())
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		predictions, err := client.ListPredictions(ctx)
		if err != nil {
			return fmt.Errorf("failed to get predictions: %w", err)
		}

		columns := []table.Column{
			{Title: "ID", Width: 20},
			{Title: "Version", Width: 20},
			{Title: "", Width: 3},
			{Title: "Created", Width: 20},
		}

		rows := []table.Row{}

		for _, prediction := range predictions.Results {
			rows = append(rows, table.Row{
				prediction.ID,
				prediction.Version,
				util.StatusSymbol(prediction.Status),
				prediction.CreatedAt,
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
