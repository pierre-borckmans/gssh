package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/gcloud"
	"gssh/views"
	"gssh/views/configurations"
	"gssh/views/instances"
	"gssh/views/statusbar"
	"os"
)

var _ tea.Model = model{}

type model struct {
	layout                bl.BubbleLayout
	configurationsPanelId bl.ID
	instancesPanelId      bl.ID
	historyPanelId        bl.ID
	statusPanelId         bl.ID
	configSize            bl.Size
	instSize              bl.Size
	historySize           bl.Size
	statusSize            bl.Size

	activePanel views.ActivePanel

	configurations tea.Model
	instances      tea.Model
	statusBar      tea.Model

	exited bool

	selectedConfiguration *gcloud.Configuration
	selectedInstance      *gcloud.Instance
}

func initialModel() model {
	layout := bl.New()
	configurationsPanelId := layout.Add("")
	instancesPanelId := layout.Add("wrap")
	historyPanelId := layout.Add("spanw 2 wrap")
	statusPanelId := layout.Add("dock south 1!")
	return model{
		layout:                layout,
		configurationsPanelId: configurationsPanelId,
		instancesPanelId:      instancesPanelId,
		historyPanelId:        historyPanelId,
		statusPanelId:         statusPanelId,
		activePanel:           views.Configurations,
		configurations:        configurations.InitialModel(),
		instances:             instances.InitialModel(),
		statusBar:             statusbar.InitialModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.configurations.Init(), m.instances.Init())
}

func (m model) updateFocus() {
	m.instances.Update(instances.BlurMsg{})
	m.configurations.Update(configurations.BlurMsg{})

	switch m.activePanel {
	case views.Configurations:
		m.configurations.Update(configurations.FocusMsg{})
	case views.Instances:
		m.instances.Update(instances.FocusMsg{})
	case views.History:
	}
	m.statusBar.Update(statusbar.SetActivePanelMsg{ActivePanel: m.activePanel})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case instances.RefreshMsg:
		_, refreshCmd := m.instances.Update(msg)
		cmds = append(cmds, refreshCmd)

	case instances.ResultMsg:
		m.instances.Update(msg)

	case instances.ErrMsg:
		m.instances.Update(msg)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.exited = true
			return m, tea.Quit

		case "left", "shift+tab":
			m.activePanel = (m.activePanel - 1 + 3) % 3
			m.updateFocus()

		case "right", "tab":
			m.activePanel = (m.activePanel + 1) % 3
			m.updateFocus()

		case "/":
			m.activePanel = views.Instances
			m.updateFocus()
			m.instances.Update(msg)

		case "r":
			_, refreshCmd := m.instances.Update(instances.RefreshMsg{
				ConfigName: m.selectedConfiguration.Name,
				ClearCache: true,
			})
			cmds = append(cmds, refreshCmd)

		default:
			switch m.activePanel {
			case views.Instances:
				_, cmd = m.instances.Update(msg)
			case views.Configurations:
				_, cmd = m.configurations.Update(msg)
			case views.History:
			}
		}

	case tea.WindowSizeMsg:
		return m, func() tea.Msg {
			return m.layout.Resize(msg.Width, msg.Height)
		}

	case bl.BubbleLayoutMsg:
		m.configSize, _ = msg.Size(m.configurationsPanelId)
		m.instSize, _ = msg.Size(m.instancesPanelId)
		m.historySize, _ = msg.Size(m.historyPanelId)
		m.statusSize, _ = msg.Size(m.statusPanelId)
		m.configurations.Update(m.configSize)
		m.instances.Update(m.instSize)
		m.statusBar.Update(m.statusSize)

	case configurations.ConfigurationSelectedMsg:
		m.selectedConfiguration = msg.Configuration

	case instances.InstanceSelectedMsg:
		m.selectedInstance = msg.Instance
		return m, tea.Quit

	default:
		switch m.activePanel {
		case views.Instances:
			_, cmd = m.instances.Update(msg)
		case views.Configurations:
			_, cmd = m.configurations.Update(msg)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinVertical(
		0, lipgloss.JoinHorizontal(
			0,
			views.BoxStyle(
				m.configSize, false).Render(m.configurations.View()),
			views.BoxStyle(
				m.instSize, false).Render(m.instances.View()),
		),
		views.BoxStyle(
			m.historySize, false).Render("History"),
		views.BoxStyle(
			m.statusSize, false).Render(m.statusBar.View()),
	)
}

func main() {
	for {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen())
		r, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
		if m, ok := r.(model); ok {
			if m.exited {
				fmt.Println("\n👋 See you soon!")
				os.Exit(0)
			}

			if m.selectedInstance != nil && m.selectedConfiguration != nil {
				fmt.Println()
				fmt.Println(lipgloss.JoinHorizontal(
					0,
					lipgloss.NewStyle().Bold(true).Render("🚀 SSHing to instance "),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#7275ff")).Render(fmt.Sprintf("[%v]", m.selectedConfiguration.Name)),
					lipgloss.NewStyle().Render(" -> "),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#ee6ff8")).Render(fmt.Sprintf("%v\n", m.selectedInstance.Name)),
					" ...",
				))
				fmt.Println()

				err = m.selectedInstance.SSH(m.selectedConfiguration.Name)
				if err != nil {
					fmt.Println(lipgloss.JoinHorizontal(
						0,
						lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff253b")).Render("Error SSHing to instance: "),
						lipgloss.NewStyle().Foreground(lipgloss.Color("#ff666b")).Render(err.Error()),
					))
					os.Exit(1)
				}
				fmt.Println("\n🛬 SSH session closed.")
			}
		}
	}
}
