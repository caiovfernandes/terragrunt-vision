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

var (
	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
	focusedStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF00FF"))

	models []tea.Model
)

const (
	initialContent string = "# Terragrunt Runner"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)
var choices = []string{"Taro", "Coffee", "Lychee"}

type views int
type status int

const (
	main views = iota
	filter
)

type Filter struct {
	region  string
	stack   string
	project string
}

type Item struct {
	title         string
	description   string
	path          string
	content       string
	lastExecution string
	cursor        int
	choice        string
	file          terragrunt.File
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	list             list.Model
	fullList         list.Model
	codeViewPort     viewport.Model
	tfViewPort       viewport.Model
	viewportRenderer *glamour.TermRenderer
	planView         bool
	focused          views
	cursor           int
	workspace        terragrunt.Workspace
	regions          []string
	projects         []string
	stacks           []string
}

func (m *Model) Init() tea.Cmd {
	m.list = m.fullList
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.focused {
	case main:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				currentItem := m.list.SelectedItem()
				item := currentItem.(Item)
				item.lastExecution = "Running terraform"
				m.list.SetItem(m.list.Index(), item)
				return m, runTerraformInit(item, m.list.Index())
			case "n":
				m.next()
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.regions) {
					m.cursor = 0
				}
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.regions) - 1
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
	case filter:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				filterCriteria := Filter{
					region: m.regions[m.cursor],
				}
				m.UpdateListItems(filterCriteria)
				m.focused = main
			case "n":
				m.next()
			case "down", "j":
				m.cursor++
				if m.cursor >= len(m.regions) {
					m.cursor = 0
				}
			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(m.regions) - 1
				}
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.focused == filter {
		s := strings.Builder{}
		s.WriteString("Account filter?\n\n")
		for i, region := range m.regions {
			if m.cursor == i {
				s.WriteString("(•) ")
			} else {
				s.WriteString("( ) ")
			}
			s.WriteString(region)
			s.WriteString("\n")
		}
		s.WriteString("\n(press q to quit)\n")

		return s.String()
	}
	if m.focused == main {
		currentItem := m.list.SelectedItem()
		var codeStr string
		var err error
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

func (m *Model) UpdateListItems(filterCriteria Filter) {
	m.list = m.fullList
	var filteredItems []list.Item

	if filterCriteria.region != "All" {
		for _, item := range m.list.Items() {
			i := item.(Item)
			//if strings.Contains(i.description, filterCriteria.project) || strings.Contains(i.description, filterCriteria.region) || strings.Contains(i.title, filterCriteria.stack) {
			if strings.Contains(i.description, filterCriteria.region) {
				filteredItems = append(filteredItems, i)
			}
		}
		m.list.SetItems(filteredItems)
	}
}

func (m *Model) next() {
	if m.focused == filter {
		m.focused = main
	} else {
		m.focused++
	}
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
	m := Model{
		fullList:         list.New(items, list.NewDefaultDelegate(), 0, 0),
		codeViewPort:     viewPortModel,
		viewportRenderer: renderer,
		tfViewPort:       viewPortModel,
		workspace:        workspace,
		regions:          append(workspace.GetRegions(), "All"),
		projects:         append(workspace.GetProjects(), "All"),
		stacks:           append(workspace.GetStacks(), "All"),
	}
	m.list.Title = "Terragrunt Files"
	fmt.Println(m.stacks)
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
