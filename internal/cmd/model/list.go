package model

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/replicate/cli/internal/util"
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
			url := fmt.Sprintf("https://replicate.com/%s", m.table.SelectedRow()[0])
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
	Short: "List models",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		models, err := util.ListModelsOnExplorePage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}

		if cmd.Flags().Changed("json") || !util.IsTTY() {
			bytes, err := json.MarshalIndent(models, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to serialize models: %w", err)
			}
			fmt.Println(string(bytes))
			return nil
		}

		if cmd.Flags().Changed("filter") {
			filter, err := cmd.Flags().GetString("filter")
			if err != nil {
				return fmt.Errorf("failed to get filter: %w", err)
			}

			pattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(filter), "\\*", ".*")
			re, err := regexp.Compile(pattern)
			if err != nil {
				return fmt.Errorf("invalid filter, %s: %w", pattern, err)
			}

			filtered := []util.ExplorePageListing{}
			for _, model := range *models {
				if re.MatchString(model.String()) {
					filtered = append(filtered, model)
				}
			}
			models = &filtered
		}

		columns := []table.Column{
			{Title: "Name", Width: 40},
			{Title: "Description", Width: 80},
			{Title: "Updated", Width: 20},
		}

		rows := []table.Row{}

		for _, model := range *models {
			rows = append(rows, table.Row{
				model.String(),
				model.Description,
				model.LatestVersionCreatedAt,
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

func init() {
	listCmd.Flags().Bool("json", false, "Print output as JSON instead of a table")
	listCmd.Flags().String("filter", "", "Filter by pattern")
}
