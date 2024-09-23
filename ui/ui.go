package ui

import (
	"fmt"
	"github.com/caiovfernandes/terragrunt-runner/terragrunt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
)

type terragruntFinishedMsg struct{ err error }

const initialContent = `
# Terragrunt Runner
`

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Item struct {
	title         string
	description   string
	path          string
	content       string
	lastExecution string
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	list             list.Model
	viewport         viewport.Model
	viewportRenderer *glamour.TermRenderer
	planView         bool
	planOutputReady  bool
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) bubbleteaExec() tea.Cmd {
	cmd := exec.Command("terragrunt", "init")

	// Set the working directory
	cmd.Dir = "/home/caio/projects/prophecy/aws-prophecy-emite-infra/workspaces/prophecy-prod/us-east-2/elasticbeanstalk/environments/sgws"

	//out, err := cmd.CombinedOutput()
	//if err != nil {
	//	log.Fatalf("command failed: %v", err)
	//}

	//currentItem := m.list.SelectedItem().(Item)
	//currentItem.lastExecution = string(out)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return terragruntFinishedMsg{err}
	})
}
func newDefaultViewPort() (viewport.Model, *glamour.TermRenderer, error) {
	vp := viewport.New(85, 27)
	vp.Style = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		PaddingRight(2)

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	str, err := renderer.Render(initialContent)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	vp.SetContent(str)
	return vp, renderer, nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "enter" {
			return m, m.bubbleteaExec()
		}
		if msg.String() == "n" {
			m.planView = true
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	// return docStyle.Render(m.list.View())
	// viewportView := viewport.Model{}

	currentItem := m.list.SelectedItem()
	var str string
	var err error
	if m.planView {
		str, err = m.viewportRenderer.Render(currentItem.(Item).lastExecution)
	} else {
		str, err = m.viewportRenderer.Render(currentItem.(Item).content)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m.viewport.SetContent(str)
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.list.View(),
		m.viewport.View(),
	)
}

func Start() {
	projects, err := terragrunt.GetProjects()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var items []list.Item
	for projectName, project := range projects {
		for regionName, region := range project.Regions {
			for stackName, stack := range region.Stacks {
				for _, file := range stack.Files {
					items = append(items, Item{
						title:         stackName,
						description:   fmt.Sprintf("Project: %s, Region: %s", projectName, regionName),
						content:       fmt.Sprintf("# `%s`\n", file.Path) + "\n```terraform\n" + file.Content + "\n```",
						path:          file.Path,
						lastExecution: "# No execution yet",
					})
				}
			}
		}
	}
	viewPortModel, renderer, err := newDefaultViewPort()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	m := Model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), viewport: viewPortModel, viewportRenderer: renderer}
	m.list.Title = "Terragrunt Files"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
