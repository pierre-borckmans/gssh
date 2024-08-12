package history

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/history"
	"gssh/views"
	"time"
)

var _ tea.Model = &Model{}

type FocusMsg struct{}
type BlurMsg struct{}
type ErrMsg struct {
	err error
}
type ResultMsg struct {
	history []*history.Connection
	items   []list.Item
}
type ClearMsg struct{}
type ConnectionSelectedMsg struct {
	Connection *history.Connection
}
type SpeedDialMsg struct {
	ConnectionIndex int
}

type Model struct {
	focused bool
	size    bl.Size
	error   error

	list        list.Model
	connections []*history.Connection
}

func RefreshHistory() tea.Msg {
	history, err := history.ListHistory()
	if err != nil {
		return ErrMsg{err}
	}
	items := make([]list.Item, 0)
	for _, connection := range history {
		items = append(items, connection)
	}
	return ResultMsg{history, items}
}

func InitialModel() *Model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.Styles.Title = l.Styles.Title.Background(lipgloss.NoColor{}).Padding(0, 0)

	return &Model{
		list: l,
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg {
		return RefreshHistory()
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FocusMsg:
		m.focused = true

	case BlurMsg:
		m.focused = false

	case ErrMsg:
		m.error = msg.err

	case ResultMsg:
		m.connections = msg.history
		m.error = nil
		m.list.SetItems(msg.items)

	case ClearMsg:
		return m, func() tea.Msg {
			history.ClearHistory()
			return RefreshHistory()
		}

	case SpeedDialMsg:
		m.list.Select(msg.ConnectionIndex)
		c, ok := m.list.SelectedItem().(*history.Connection)
		if !ok {
			return m, nil
		}
		return m, func() tea.Msg {
			time.Sleep(time.Millisecond * 500)
			return ConnectionSelectedMsg{c}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			c, ok := m.list.SelectedItem().(*history.Connection)
			if !ok {
				return m, nil
			}

			return m, func() tea.Msg {
				return ConnectionSelectedMsg{c}
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
			titleStyle.Foreground(lipgloss.Color("#ffffff")).Render(" Connection history "),
			titleStyle.Render(" "),
		),
	)

	if m.error != nil {
		return style.Align(lipgloss.Center, lipgloss.Center).Foreground(lipgloss.Color("202")).Render(
			"Error getting history",
		)
	}

	return style.Render(
		m.list.View(),
	)
}
