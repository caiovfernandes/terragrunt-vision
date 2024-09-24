package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/caiovfernandes/terragrunt-runner/terragrunt"
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
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

func main() {

	projects, err := terragrunt.GetProjects()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	columns := []table.Column{
		{Title: "Resource", Width: 30},
		{Title: "Stack", Width: 10},
		{Title: "Region", Width: 10},
		{Title: "Account", Width: 20},
	}

	var rows []table.Row
	for projectName, project := range projects {
		for regionName, region := range project.Regions {
			for stackName, stack := range region.Stacks {
				for _, file := range stack.Files {
					rows = append(rows, table.Row{
						func() string {
							parts := strings.Split(file.Path, "/")
							if len(parts) >= 3 {
								return parts[len(parts)-2]
							}
							return "Default"
						}(),
						stackName,
						regionName,
						projectName,
					})
				}
			}
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
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
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
