package ui

//type status int
//
//var (
//	columnStyle  = lipgloss.NewStyle().Padding(1, 2)
//	focusedStyle = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#FF00FF"))
//
//	models []tea.Model
//)
//
//const (
//	account status = iota
//	region
//	stack
//)
//
//const (
//	model status = iota
//	form
//)
//
//// Item
//
//type Item struct {
//	status      status
//	title       string
//	description string
//}
//
//func (t Item) FilterValue() string {
//	return t.title
//}
//
//func (t Item) Title() string {
//	return t.title
//}
//
//func (t Item) Description() string {
//	return t.description
//}
//
//func (t *Item) Next() {
//	if t.status == stack {
//		t.status = account
//	} else {
//		t.status++
//	}
//}
//
//func NewTask(status status, title, description string) Item {
//	return Item{status: status, title: title, description: description}
//}
//
//// Model
//
//type Model struct {
//	focused status
//	lists   []list.Model
//	err     error
//}
//
//func New() *Model {
//	return &Model{}
//}
//
//func (m *Model) MoveToNext() {
//	selectedItem := m.lists[m.focused].SelectedItem()
//	selectedTask := selectedItem.(Item)
//	m.lists[selectedTask.status].RemoveItem(m.lists[m.focused].Index())
//	selectedTask.Next()
//	m.lists[selectedTask.status].InsertItem(len(m.lists[selectedTask.status].Items())-1, selectedTask)
//}
//
//func (m *Model) initLists(width, height int) {
//	defaultList := list.New(
//		[]list.Item{},
//		list.NewDefaultDelegate(),
//		width,
//		height-5,
//	)
//	defaultList.SetShowHelp(false)
//	m.lists = []list.Model{defaultList, defaultList, defaultList}
//
//	// Account
//	m.lists[account].Title = "To Do"
//	m.lists[account].SetItems([]list.Item{
//		Item{status: account, title: "Write documentation", description: "Write documentation for the project"},
//		Item{status: account, title: "Write tests", description: "Write tests for the project"},
//		Item{status: account, title: "Write code", description: "Write code for the project"},
//	})
//
//	// Regions
//	m.lists[region].Title = "In Progress"
//	m.lists[region].SetItems([]list.Item{
//		Item{status: region, title: "SoW", description: "Write statement of work."},
//		Item{status: region, title: "Leverage Requirements", description: "Leverage project functional and non-functional requirements."},
//		Item{status: region, title: "Architectural Documentation", description: "Write documentation about the solution architecture."},
//	})
//
//	// Stacks
//	m.lists[stack].Title = "Done"
//	m.lists[stack].SetItems([]list.Item{
//		Item{status: stack, title: "Project Idea", description: "That's the initial idea to creat this project......"},
//	})
//}
//
//func (m *Model) Init() tea.Cmd {
//	m.focused = 0
//	m.lists = make([]list.Model, 3)
//	return nil
//}
//
//func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	switch msg := msg.(type) {
//	case tea.WindowSizeMsg:
//		m.initLists(msg.Width, msg.Height/2)
//	case tea.KeyMsg:
//		switch msg.String() {
//		case "h", "left":
//			if m.focused > account {
//				m.focused--
//			}
//		case "l", "right":
//			if m.focused < stack {
//				m.focused++
//			}
//		case "enter":
//			m.MoveToNext()
//		case "n":
//			models[model] = m
//			models[form].(*Form).focused = m.focused
//			return models[form], nil
//		case "ctrl+c", "q":
//			return m, tea.Quit
//		}
//	case Item:
//		task := msg
//		return m, m.lists[task.status].InsertItem(len(m.lists[task.status].Items()), task)
//	}
//
//	var cmd tea.Cmd
//	m.lists[m.focused], cmd = m.lists[m.focused].Update(msg)
//	return m, cmd
//}
//
//func (m *Model) View() string {
//	return lipgloss.JoinHorizontal(
//		lipgloss.Left,
//		getBoardStyle(m, account, m.lists[account]),
//		getBoardStyle(m, region, m.lists[region]),
//		getBoardStyle(m, stack, m.lists[stack]),
//	)
//}
//
//type Form struct {
//	focused     status
//	title       textinput.Model
//	description textarea.Model
//}
//
//func NewForm(focused status) *Form {
//	form := &Form{focused: focused}
//	form.title = textinput.New()
//	form.title.Focus()
//	form.description = textarea.New()
//	return form
//}
//
//func (f Form) CreateTaskFromForm(m *Model) tea.Msg {
//	m.lists[f.focused].InsertItem(len(m.lists[f.focused].Items()), NewTask(f.focused, f.title.Value(), f.description.Value()))
//	return NewTask(f.focused, f.title.Value(), f.description.Value())
//}
//
//func (f Form) Init() tea.Cmd {
//	return nil
//}
//
//func (f Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
//	var cmd tea.Cmd
//	switch msg := msg.(type) {
//	case tea.KeyMsg:
//		switch msg.String() {
//		case "enter":
//			if f.title.Focused() {
//				f.title.Blur()
//				f.description.Focus()
//				return f, textarea.Blink
//			} else {
//				models[form] = f
//				f.CreateTaskFromForm(models[model].(*Model))
//				models[form] = NewForm(f.focused)
//				return models[model], nil
//			}
//		}
//	}
//	if f.title.Focused() {
//		f.title, cmd = f.title.Update(msg)
//		return f, cmd
//	} else {
//		f.description, cmd = f.description.Update(msg)
//		return f, cmd
//	}
//}
//
//func (f Form) View() string {
//	return lipgloss.JoinVertical(lipgloss.Left, f.title.View(), f.description.View())
//}
//
//// Functions
//func getBoardStyle(m *Model, s status, listItem list.Model) string {
//	if s == m.focused {
//		return focusedStyle.Render(listItem.View())
//	} else {
//		return columnStyle.Render(listItem.View())
//	}
//}
//
//func main() {
//	models = []tea.Model{New(), NewForm(account)}
//	m := models[model]
//	m.Init()
//
//	p := tea.NewProgram(m)
//	if _, err := p.Run(); err != nil {
//		fmt.Println("Error starting program:", err)
//		panic(err)
//	}
//}
