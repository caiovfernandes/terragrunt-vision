package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/caiovfernandes/terragrunt-runner/terragrunt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type status int

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF00FF"))

	models []tea.Model
)

const (
	initialContent string = "# Terragrunt Runner"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type Item struct {
	title         string
	description   string
	path          string
	content       string
	lastExecution string
	cursor        int
	choice        string
}

type views int

const (
	filter views = iota
	main
)

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	list             list.Model
	codeViewPort     viewport.Model
	tfViewPort       viewport.Model
	viewportRenderer *glamour.TermRenderer
	planView         bool
	focused          views
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func newDefaultViewPort() (viewport.Model, *glamour.TermRenderer, error) {
	vp := viewport.New(100, 27)
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
			currentItem := m.list.SelectedItem()
			item := currentItem.(Item)
			item.lastExecution = "Running terraform"
			m.list.SetItem(m.list.Index(), item)

			return m, runTerraformInit(item, m.list.Index())
		}
		if msg.String() == "n" {
			if m.planView {
				m.planView = false
			} else {
				m.planView = true
			}
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width, msg.Height-v)
		m.codeViewPort.Height = msg.Height - h
		m.tfViewPort.Height = msg.Height - h
	case terraformInitMsg:
		item := m.list.Items()[msg.Index].(Item)
		item.lastExecution = "# Output:\n\n```shell" + msg.Output + "```\n # StdErr:\n " + msg.Error.Error() // Assuming we add a method to set this value
		m.list.SetItem(msg.Index, item)
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {

	if m.focused == filter {
		s := strings.Builder{}
		s.WriteString("What kind of Bubble Tea would you like to order?\n\n")

		for i := 0; i < len(choices); i++ {
			if m.cursor == i {
				s.WriteString("(•) ")
			} else {
				s.WriteString("( ) ")
			}
			s.WriteString(choices[i])
			s.WriteString("\n")
		}
		s.WriteString("\n(press q to quit)\n")

		return s.String()
	}
	if m.focused == main {

		currentItem := m.list.SelectedItem()
		var codeStr string
		var err error
		//if m.planView {
		//	codeStr, err = m.viewportRenderer.Render(string(currentItem.(Item).lastExecution))
		//} else {
		//	codeStr, err = m.viewportRenderer.Render(currentItem.(Item).content)
		//}

		codeStr, err = m.viewportRenderer.Render(currentItem.(Item).content)
		tfRunStr, err := m.viewportRenderer.Render(currentItem.(Item).lastExecution)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		m.codeViewPort.SetContent(codeStr)
		m.tfViewPort.SetContent(tfRunStr)
		return lipgloss.JoinHorizontal(
			lipgloss.Center,
			m.list.View(),
			m.codeViewPort.View(),
			m.tfViewPort.View(),
		)
	}
	return ""
}

func Start() {
	workspace, err := terragrunt.GetWorkspace()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var items []list.Item
	for projectName, project := range workspace.Projects {
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
	m := Model{list: list.New(items, list.NewDefaultDelegate(), 0, 0), codeViewPort: viewPortModel, viewportRenderer: renderer, tfViewPort: viewPortModel}
	m.list.Title = "Terragrunt Files"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type terraformInitMsg struct {
	Output string
	Item   Item
	Index  int
	Error  error
}

func runTerraformInit(item Item, itemPosition int) tea.Cmd {
	return func() tea.Msg {
		output, err := terragrunt.RunTerraformInit(item.path)
		return terraformInitMsg{Output: output, Item: item, Index: itemPosition, Error: err}
	}
}
