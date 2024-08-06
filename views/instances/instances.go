package instances

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gssh/gcloud"
)

var docStyle = lipgloss.NewStyle().Padding(2, 1)

var _ tea.Model = &Model{}

type ActivateMsg struct{}
type DeactivateMsg struct{}

var Activate tea.Msg = ActivateMsg{}
var Deactivate tea.Msg = DeactivateMsg{}

type Model struct {
	list             list.Model
	error            error
	instances        []*gcloud.Instance
	selectedInstance *gcloud.Instance
	active           bool
}

func InitialModel() *Model {
	instances, err := gcloud.ListInstances()
	if err != nil {
		return &Model{
			error: err,
		}
	}

	items := make([]list.Item, 0)
	for i, inst := range instances {
		if inst.Status == gcloud.InstanceStatusRunning {
			items = append(items, inst)
		}
		if i > 2 {
			break
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a GCP instance to SSH into:"

	return &Model{
		list:      l,
		instances: instances,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ActivateMsg:
		m.active = true

	case DeactivateMsg:
		m.active = false
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			i, ok := m.list.SelectedItem().(*gcloud.Instance)
			if ok {
				m.selectedInstance = i
			}
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
	style := lipgloss.NewStyle().Padding(0, 2).Border(lipgloss.RoundedBorder())
	selectedStyle := style.BorderForeground(lipgloss.Color("#5f5fd7"))

	if m.active {
		style = selectedStyle
		m.list.Styles.Title = m.list.Styles.Title.Background(lipgloss.Color("62"))
	} else {
		m.list.Styles.Title = m.list.Styles.Title.Background(lipgloss.NoColor{})
	}

	m.list.SetShowStatusBar(false)
	m.list.SetShowHelp(false)
	return style.Render(m.list.View())
}
