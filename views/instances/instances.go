package instances

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/gcloud"
	"gssh/views"
)

var _ tea.Model = &Model{}

type FocusMsg struct{}
type BlurMsg struct{}
type RefreshMsg struct {
	ConfigName string
	ClearCache bool
}

type ErrMsg struct {
	err error
}
type ResultMsg struct {
	instances []*gcloud.Instance
	items     []list.Item
}
type InstanceSelectedMsg struct {
	Instance *gcloud.Instance
}

type Model struct {
	focused bool
	size    bl.Size
	loading bool
	error   error

	configName       string
	list             list.Model
	instances        []*gcloud.Instance
	selectedInstance *gcloud.Instance
}

func InitialModel() *Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(true)
	l.Styles.Title = l.Styles.Title.Background(lipgloss.NoColor{}).Padding(0, 0)
	l.FilterInput.Prompt = "🔍 "
	l.FilterInput.Placeholder = "Filter instances..."

	return &Model{
		list:    l,
		loading: true,
	}
}

func RefreshInstances(configName string, clearCache bool) tea.Msg {
	instances, err := gcloud.ListInstances(configName, clearCache)
	if err != nil {
		return ErrMsg{err}
	}

	items := make([]list.Item, 0)
	for _, inst := range instances {
		if inst.Status == gcloud.InstanceStatusRunning {
			items = append(items, inst)
		}
	}
	return ResultMsg{instances, items}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FocusMsg:
		m.focused = true

	case BlurMsg:
		m.focused = false

	case RefreshMsg:
		m.loading = true
		m.configName = msg.ConfigName
		return m, func() tea.Msg {
			return RefreshInstances(msg.ConfigName, msg.ClearCache)
		}

	case ErrMsg:
		m.loading = false
		m.error = msg.err

	case ResultMsg:
		m.instances = msg.instances
		m.loading = false
		m.error = nil
		m.list.SetItems(msg.items)

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			i, ok := m.list.SelectedItem().(*gcloud.Instance)
			if ok {
				m.selectedInstance = i
			}

			if m.list.FilterState() != list.Filtering {
				return m, func() tea.Msg {
					return InstanceSelectedMsg{m.selectedInstance}
				}
			}
		case "esc":
			if m.list.FilterState() != list.Filtering {
				return m, nil
			}
		}

	case bl.Size:
		x, y := views.PanelStyle.GetFrameSize()
		m.size = msg
		m.list.SetSize(msg.Width-x-2, msg.Height-y-2)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	style := views.PanelStyle.Width(m.size.Width - 2).Height(m.size.Height - 2)
	selectedStyle := style.BorderForeground(lipgloss.Color("#5f5fd7"))

	configStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ee6ff8"))

	filterStr := ""
	if m.list.FilterValue() != "" {
		filterStr = lipgloss.NewStyle().
			Background(lipgloss.Color("#baa000")).
			Foreground(lipgloss.Color("#ffffff")).
			Render(fmt.Sprintf(" 🔍 \"%v\" ", m.list.FilterValue()))
	}

	titleStyle := lipgloss.NewStyle()

	if m.focused {
		style = selectedStyle
		titleStyle = titleStyle.Background(lipgloss.Color("62"))
		configStyle = configStyle.Background(lipgloss.Color("62"))
	} else {
		titleStyle = titleStyle.Background(lipgloss.NoColor{})
		configStyle = configStyle.Background(lipgloss.NoColor{})
	}

	m.list.Title = lipgloss.JoinHorizontal(0,
		lipgloss.JoinHorizontal(0,
			titleStyle.Foreground(lipgloss.Color("#ffffff")).Render(" Select a GCP instance in "),
			configStyle.Render(fmt.Sprintf("[%v]", m.configName)),
			titleStyle.Render(" "),
		),
		" ",
		filterStr,
	)

	if m.error != nil {
		return style.Align(lipgloss.Center, lipgloss.Center).Foreground(lipgloss.Color("202")).Render(
			fmt.Sprintf("Error fetching instances for [%v]\n%v", m.configName, m.error.Error()),
		)
	}

	if m.loading {
		return style.Align(lipgloss.Center, lipgloss.Center).Render(fmt.Sprintf("Fetching instances for %s...", m.configName))
	}

	return style.Render(m.list.View())
}
