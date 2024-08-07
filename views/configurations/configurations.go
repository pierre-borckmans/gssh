package configurations

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/gcloud"
	"gssh/views"
	"gssh/views/instances"
)

var _ tea.Model = &Model{}

type FocusMsg struct{}
type BlurMsg struct{}
type ConfigurationSelectedMsg struct{ Configuration *gcloud.Configuration }
type RefreshMsg struct{}

type ErrMsg struct {
	err error
}
type ResultMsg struct {
	configurations []*gcloud.Configuration
	items          []list.Item
	activeIdx      int
}
type Model struct {
	size           bl.Size
	list           list.Model
	error          error
	configurations []*gcloud.Configuration
	focused        bool
}

func InitialModel() *Model {
	configs, err := gcloud.ListConfigurations()
	if err != nil {
		return &Model{
			error: err,
		}
	}

	var activeConfigIdx int
	items := make([]list.Item, 0)
	for i, config := range configs {
		items = append(items, config)
		if config.Active {
			activeConfigIdx = i
		}
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a GCP configuration:"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.Select(activeConfigIdx)

	return &Model{
		list:           l,
		configurations: configs,
		focused:        true,
	}
}

func RefreshConfigurations() tea.Msg {
	configs, err := gcloud.ListConfigurations()
	if err != nil {
		return ErrMsg{err}
	}
	items := make([]list.Item, 0)
	var activeConfigIdx int
	for i, config := range configs {
		items = append(items, config)
		if config.Active {
			activeConfigIdx = i
		}

	}
	return ResultMsg{configs, items, activeConfigIdx}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			return ConfigurationSelectedMsg{
				Configuration: m.list.SelectedItem().(*gcloud.Configuration),
			}
		},
		func() tea.Msg {
			return instances.RefreshMsg{
				ConfigName: m.list.SelectedItem().(*gcloud.Configuration).Name,
			}
		},
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FocusMsg:
		m.focused = true
	case BlurMsg:
		m.focused = false

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil
		case "enter":
			return m, func() tea.Msg {
				if err := gcloud.ActivateConfiguration(m.list.SelectedItem().(*gcloud.Configuration).Name); err != nil {
					return ErrMsg{err}
				}
				return RefreshMsg{}
			}
		}

	case bl.Size:
		x, y := views.PanelStyle.GetFrameSize()
		m.size = msg
		m.list.SetSize(msg.Width-x, msg.Height-y-2)

	case RefreshMsg:
		return m, func() tea.Msg {
			return RefreshConfigurations()
		}

	case ResultMsg:
		m.configurations = msg.configurations
		m.list.SetItems(msg.items)
		m.list.Select(msg.activeIdx)
	}

	var cmds []tea.Cmd
	newList, cmd := m.list.Update(msg)
	cmds = append(cmds, cmd)
	if newList.SelectedItem() != m.list.SelectedItem() {
		cmds = append(cmds, func() tea.Msg {
			return instances.RefreshMsg{
				ConfigName: m.list.SelectedItem().(*gcloud.Configuration).Name,
			}
		})
	}
	m.list = newList
	cmds = append(cmds, func() tea.Msg {
		return ConfigurationSelectedMsg{
			Configuration: m.list.SelectedItem().(*gcloud.Configuration),
		}
	})
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	style := views.PanelStyle.Width(m.size.Width - 2).Height(m.size.Height - 2)
	selectedStyle := style.BorderForeground(lipgloss.Color("#5f5fd7"))

	if m.focused {
		style = selectedStyle
		m.list.Styles.Title = m.list.Styles.Title.Background(lipgloss.Color("62"))
	} else {
		m.list.Styles.Title = m.list.Styles.Title.Background(lipgloss.NoColor{})
	}

	return style.Render(m.list.View())
}
