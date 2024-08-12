package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/views"
)

var _ tea.Model = &Model{}
var baseStyle = lipgloss.NewStyle().Background(lipgloss.Color("62"))

type SetActivePanelMsg struct {
	ActivePanel views.ActivePanel
}

type Model struct {
	size        bl.Size
	activePanel views.ActivePanel
}

func InitialModel() *Model {
	return &Model{
		activePanel: views.Configurations,
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case bl.Size:
		m.size = msg
	case SetActivePanelMsg:
		m.activePanel = msg.ActivePanel
	}

	return m, nil
}

func shortcut(key string, action string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		baseStyle.Background(lipgloss.Color("5")).Foreground(lipgloss.Color("62")).Render("▌"+key+"▐"),
		baseStyle.Foreground(lipgloss.Color("#bbbbbb")).Render(action),
		baseStyle.Render(" "),
	)
}

func (m *Model) View() string {
	var activeView string

	var enter string
	var arrows string
	switch m.activePanel {
	case views.Configurations:
		activeView = "Configurations"
		enter = "Activate configuration"
		arrows = "Browse configurations"
	case views.Instances:
		activeView = "Instances"
		enter = "SSH to instance"
		arrows = "Browse instances"
	case views.History:
		activeView = "History"
		enter = "SSH to instance"
		arrows = "Browse history"
	default:
		enter = ""
	}

	shortcuts := baseStyle.Padding(0, 1).Align(lipgloss.Right, lipgloss.Center).Render(
		lipgloss.JoinHorizontal(
			0,
			shortcut("↑↓", arrows),
			shortcut("⇥", "Next panel"),
			shortcut("/", "Filter instances"),
			shortcut("↵", enter),
			shortcut("R", "Reload instances"),
			shortcut("C", "Clear history"),
			shortcut("Q", "Quit"),
		),
	)

	return baseStyle.Width(m.size.Width).Padding(0, 1).Render(
		lipgloss.JoinHorizontal(
			0,
			baseStyle.Width(m.size.Width-lipgloss.Width(shortcuts)).Padding(0, 1).Foreground(lipgloss.Color("4")).Render("["+activeView+"]"),
			shortcuts,
		),
	)
}
