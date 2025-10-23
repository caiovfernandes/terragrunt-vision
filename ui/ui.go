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

const (
	initialContent string = "# Terragrunt Runner\n\nSelect items with **space**, execute with **enter**\n\nAvailable commands:\n- **i** - init\n- **p** - plan\n- **a** - apply\n- **v** - validate\n- **d** - destroy"
)

var (
	docStyle       = lipgloss.NewStyle().Margin(1, 2)
	selectedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	commandStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	progressStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type (
	views  int
	status int
)

const (
	main views = iota
	filter
	executing
)

type Filter struct {
	region  string
	stack   string
	project string
}

type ExecutionStatus int

const (
	Pending ExecutionStatus = iota
	Running
	Success
	Failed
)

type Item struct {
	title           string
	description     string
	path            string
	content         string
	lastExecution   string
	selected        bool
	executionStatus ExecutionStatus
	file            terragrunt.File
}

func (i Item) Title() string {
	prefix := "[ ]"
	if i.selected {
		prefix = selectedStyle.Render("[✓]")
	}

	statusIcon := ""
	switch i.executionStatus {
	case Running:
		statusIcon = progressStyle.Render(" ⟳")
	case Success:
		statusIcon = successStyle.Render(" ✓")
	case Failed:
		statusIcon = errorStyle.Render(" ✗")
	}

	return fmt.Sprintf("%s %s%s", prefix, i.title, statusIcon)
}

func (i Item) Description() string { return i.description }
func (i Item) FilterValue() string { return i.title }

type TerragruntCommand string

const (
	CommandInit     TerragruntCommand = "init"
	CommandPlan     TerragruntCommand = "plan"
	CommandApply    TerragruntCommand = "apply"
	CommandValidate TerragruntCommand = "validate"
	CommandDestroy  TerragruntCommand = "destroy"
)

type Model struct {
	list              list.Model
	fullList          list.Model
	codeViewPort      viewport.Model
	tfViewPort        viewport.Model
	viewportRenderer  *glamour.TermRenderer
	focused           views
	cursor            int
	workspace         terragrunt.Workspace
	regions           []string
	projects          []string
	stacks            []string
	currentCommand    TerragruntCommand
	executing         bool
	executionResults  map[int]string
	pendingExecutions int

	windowSize tea.WindowSizeMsg
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
			if m.executing {
				// Only allow quit during execution
				if msg.String() == "ctrl+c" {
					return m, tea.Quit
				}
				return m, nil
			}

			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case " ":
				// Toggle selection
				idx := m.list.Index()
				currentItem := m.list.SelectedItem().(Item)
				currentItem.selected = !currentItem.selected
				m.list.SetItem(idx, currentItem)
				return m, nil
			case "i":
				m.currentCommand = CommandInit
				return m, nil
			case "p":
				m.currentCommand = CommandPlan
				return m, nil
			case "a":
				m.currentCommand = CommandApply
				return m, nil
			case "v":
				m.currentCommand = CommandValidate
				return m, nil
			case "d":
				m.currentCommand = CommandDestroy
				return m, nil
			case "enter":
				// Execute selected command on selected items
				var selectedItems []Item
				var selectedIndices []int
				for idx, item := range m.list.Items() {
					if item.(Item).selected {
						selectedItems = append(selectedItems, item.(Item))
						selectedIndices = append(selectedIndices, idx)
					}
				}

				if len(selectedItems) == 0 {
					// If nothing selected, execute on current item
					currentItem := m.list.SelectedItem().(Item)
					selectedItems = append(selectedItems, currentItem)
					selectedIndices = append(selectedIndices, m.list.Index())
				}

				// Mark all as running
				for _, idx := range selectedIndices {
					item := m.list.Items()[idx].(Item)
					item.executionStatus = Running
					item.lastExecution = "# Running " + string(m.currentCommand) + "..."
					m.list.SetItem(idx, item)
				}

				m.executing = true
				m.pendingExecutions = len(selectedItems)
				return m, runTerragruntParallel(selectedItems, selectedIndices, m.currentCommand)
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
			h, _ := docStyle.GetFrameSize()
			m.windowSize = msg
			m.list.SetSize(msg.Width/3, msg.Height)
			m.codeViewPort.Width = msg.Width / 3
			m.codeViewPort.Height = msg.Height - h
			m.tfViewPort.Width = msg.Width / 3
			m.tfViewPort.Height = msg.Height - h
		case terragruntResultMsg:
			// Update individual item
			item := m.list.Items()[msg.Index].(Item)
			if msg.Error != nil {
				item.executionStatus = Failed
				item.lastExecution = fmt.Sprintf("# Error:\n\n```\n%s\n```", msg.Error.Error())
			} else {
				item.executionStatus = Success
				item.lastExecution = fmt.Sprintf("# Output:\n\n```shell\n%s\n```", msg.Output)
			}
			item.selected = false
			m.list.SetItem(msg.Index, item)
			m.tfViewPort.GotoBottom()

			// Decrement pending executions
			m.pendingExecutions--
			if m.pendingExecutions <= 0 {
				m.executing = false
			}
		case allExecutionsCompleteMsg:
			m.executing = false
		}
	case filter:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				filterCriteria := Filter{
					region:  m.regions[m.cursor],
					project: "",
					stack:   "",
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
		s.WriteString("Region filter\n\n")
		for i, region := range m.regions {
			if m.cursor == i {
				s.WriteString("(•) ")
			} else {
				s.WriteString("( ) ")
			}
			s.WriteString(region)
			s.WriteString("\n")
		}
		s.WriteString("\n(press enter to apply, n to cancel, q to quit)\n")

		return lipgloss.PlaceHorizontal(50, lipgloss.Center, s.String())
	}
	if m.focused == main {
		if m.isWindowSizeSet() {
			m.list.SetSize(m.windowSize.Width/3, m.windowSize.Height-4)
		}

		// Status bar with current command and execution status
		statusBar := ""
		if m.currentCommand != "" {
			statusBar = commandStyle.Render(fmt.Sprintf("Command: %s", m.currentCommand))
		} else {
			statusBar = "Command: " + commandStyle.Render("init (default)")
		}

		if m.executing {
			statusBar += " | " + progressStyle.Render("EXECUTING...")
		}

		selectedCount := 0
		for _, item := range m.list.Items() {
			if item.(Item).selected {
				selectedCount++
			}
		}
		if selectedCount > 0 {
			statusBar += fmt.Sprintf(" | Selected: %s%d%s", selectedStyle.Render(""), selectedCount, selectedStyle.Render(""))
		}

		statusBar += "\n" + lipgloss.NewStyle().Faint(true).Render("space: select | i/p/a/v/d: command | enter: execute | n: filter | q: quit")

		currentItem := m.list.SelectedItem()
		var codeStr string
		var err error
		codeStr, err = m.viewportRenderer.Render(currentItem.(Item).content)

		tfRunStr, err := m.viewportRenderer.Render(currentItem.(Item).lastExecution)
		if err != nil {
			// Don't exit on render errors, just show error
			tfRunStr = fmt.Sprintf("Error rendering: %v", err)
		}
		m.codeViewPort.SetContent(codeStr)
		m.tfViewPort.SetContent(tfRunStr)
		m.tfViewPort.GotoBottom()

		mainView := lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.list.View(),
			m.codeViewPort.View(),
			m.tfViewPort.View(),
		)

		return lipgloss.JoinVertical(lipgloss.Left, statusBar, mainView)
	}
	return ""
}

func (m *Model) UpdateListItems(filterCriteria Filter) {
	m.list = m.fullList
	var filteredItems []list.Item

	applyFilter := filterCriteria.region != "All" || filterCriteria.project != "All" || filterCriteria.stack != "All"

	if applyFilter {
		for _, item := range m.fullList.Items() {
			i := item.(Item)
			matchesRegion := filterCriteria.region == "All" || strings.Contains(i.description, filterCriteria.region)
			matchesProject := filterCriteria.project == "All" || strings.Contains(i.description, filterCriteria.project)
			matchesStack := filterCriteria.stack == "All" || strings.Contains(i.title, filterCriteria.stack)

			if matchesRegion && matchesProject && matchesStack {
				filteredItems = append(filteredItems, i)
			}
		}
		m.list.SetItems(filteredItems)
	} else {
		// Reset to full list if all filters are "All"
		m.list = m.fullList
	}
}

func (m *Model) next() {
	if m.focused == filter {
		m.focused = main
	} else {
		m.focused++
	}
}

func (m *Model) isWindowSizeSet() bool {
	return m.windowSize.Width != 0 && m.windowSize.Height != 0
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
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	var items []list.Item
	for projectName, project := range workspace.Projects {
		for regionName, region := range project.Regions {
			for stackName, stack := range region.Stacks {
				for _, file := range stack.Files {
					items = append(items, Item{
						title:           stackName,
						description:     fmt.Sprintf("Project: %s, Region: %s", projectName, regionName),
						content:         fmt.Sprintf("# `%s`\n", file.Path) + "\n```terraform\n" + file.Content + "\n```",
						path:            file.Path,
						lastExecution:   "# No execution yet",
						selected:        false,
						executionStatus: Pending,
					})
				}
			}
		}
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "No terragrunt files found in workspace\n")
		os.Exit(1)
	}

	viewPortModel, renderer, err := newDefaultViewPort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating viewport: %v\n", err)
		os.Exit(1)
	}

	tfViewPort, _, err := newDefaultViewPort()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating viewport: %v\n", err)
		os.Exit(1)
	}

	m := Model{
		fullList:         list.New(items, list.NewDefaultDelegate(), 0, 0),
		codeViewPort:     viewPortModel,
		viewportRenderer: renderer,
		tfViewPort:       tfViewPort,
		workspace:        workspace,
		regions:          append(workspace.GetRegions(), "All"),
		projects:         append(workspace.GetProjects(), "All"),
		stacks:           append(workspace.GetStacks(), "All"),
		currentCommand:   CommandInit,
		executing:        false,
		executionResults: make(map[int]string),
	}

	m.list.Title = "Terragrunt Files"
	p := tea.NewProgram(&m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

type terragruntResultMsg struct {
	Output string
	Index  int
	Error  error
}

type allExecutionsCompleteMsg struct{}

func runTerragruntParallel(items []Item, indices []int, command TerragruntCommand) tea.Cmd {
	return func() tea.Msg {
		// Create a channel to send results back to the UI
		resultsChan := make(chan terragruntResultMsg)

		// Launch all executions in parallel
		for i, item := range items {
			go func(item Item, idx int) {
				output, err := terragrunt.RunTerragruntCommand(item.path, string(command))
				resultsChan <- terragruntResultMsg{
					Output: output,
					Index:  indices[idx],
					Error:  err,
				}
			}(item, i)
		}

		// Start a goroutine to listen for all results
		// This creates a subscription that Bubble Tea will handle
		return tea.Batch(listenForResults(resultsChan, len(items))...)
	}
}

func listenForResults(results chan terragruntResultMsg, count int) []tea.Cmd {
	cmds := make([]tea.Cmd, count)
	for i := 0; i < count; i++ {
		cmds[i] = func() tea.Msg {
			return <-results
		}
	}
	return cmds
}
