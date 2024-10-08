package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
	"gssh/config"
	"gssh/gcloud"
	"gssh/history"
	"gssh/views"
	"gssh/views/configurations"
	hist_view "gssh/views/history"
	"gssh/views/instances"
	"gssh/views/statusbar"
	"os"
	"time"
)

var _ tea.Model = &model{}

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
	history        tea.Model
	statusBar      tea.Model

	filtering bool
	exited    bool

	selectedConfiguration     *gcloud.Configuration
	selectedInstance          *gcloud.Instance
	selectedHistoryConnection *history.Connection
}

func initialModel() *model {
	layout := bl.New()
	configurationsPanelId := layout.Add("")
	instancesPanelId := layout.Add("wrap")
	historyPanelId := layout.Add("spanw 2 wrap")
	statusPanelId := layout.Add("dock south 1!")
	m := &model{
		layout:                layout,
		configurationsPanelId: configurationsPanelId,
		instancesPanelId:      instancesPanelId,
		historyPanelId:        historyPanelId,
		statusPanelId:         statusPanelId,
		activePanel:           views.Configurations,
		configurations:        configurations.InitialModel(),
		instances:             instances.InitialModel(),
		history:               hist_view.InitialModel(),
		statusBar:             statusbar.InitialModel(),
	}
	return m
}

type pollTickMsg struct{}

func (m *model) pollTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return pollTickMsg{}
	})
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.configurations.Init(),
		m.instances.Init(),
		m.history.Init(),
		m.pollTick(),
	)
}

func (m *model) updateFocus() {
	m.instances.Update(instances.BlurMsg{})
	m.configurations.Update(configurations.BlurMsg{})
	m.history.Update(hist_view.BlurMsg{})

	switch m.activePanel {
	case views.Configurations:
		m.configurations.Update(configurations.FocusMsg{})
	case views.Instances:
		m.instances.Update(instances.FocusMsg{})
	case views.History:
		m.history.Update(hist_view.FocusMsg{})
	}
	m.statusBar.Update(statusbar.SetActivePanelMsg{ActivePanel: m.activePanel})
}

func (m *model) speedDial(msg tea.Msg, index int) tea.Cmd {
	if m.filtering {
		switch m.activePanel {
		case views.Instances:
			_, cmd := m.instances.Update(msg)
			return cmd
		default:
			return nil
		}
	}

	m.activePanel = views.History
	m.updateFocus()
	_, speedDialCmd := m.history.Update(hist_view.SpeedDialMsg{ConnectionIndex: index})
	return speedDialCmd
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case pollTickMsg:
		if !m.filtering {
			_, refreshInstancesCmd := m.instances.Update(instances.RefreshMsg{
				ConfigName: m.selectedConfiguration.Name,
				ClearCache: false,
			})
			_, refreshHistoryCmd := m.history.Update(hist_view.RefreshMsg{})
			cmds = append(cmds, refreshInstancesCmd)
			cmds = append(cmds, refreshHistoryCmd)
			cmds = append(cmds, m.pollTick())
		}

	case instances.RefreshMsg:
		_, refreshCmd := m.instances.Update(msg)
		cmds = append(cmds, refreshCmd)

	case instances.ResultMsg:
		m.instances.Update(msg)

	case instances.ErrMsg:
		m.instances.Update(msg)

	case instances.FilteringStateMsg:
		m.filtering = msg.Filtering

	case hist_view.ResultMsg:
		m.history.Update(msg)

	case hist_view.ErrMsg:
		m.history.Update(msg)

	case tea.KeyMsg:
		if m.filtering {
			switch m.activePanel {
			case views.Instances:
				_, cmd = m.instances.Update(msg)
				return m, cmd
			}
		}

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
			m.filtering = true
			m.activePanel = views.Instances
			m.updateFocus()
			m.instances.Update(msg)

		case "r":
			_, refreshCmd := m.instances.Update(instances.RefreshMsg{
				ConfigName: m.selectedConfiguration.Name,
				ClearCache: true,
			})
			cmds = append(cmds, refreshCmd)

		case "c":
			_, clearCmd := m.history.Update(hist_view.ClearMsg{})
			cmds = append(cmds, clearCmd)

		case "0":
			cmds = append(cmds, m.speedDial(msg, 0))
		case "1":
			cmds = append(cmds, m.speedDial(msg, 1))
		case "2":
			cmds = append(cmds, m.speedDial(msg, 2))
		case "3":
			cmds = append(cmds, m.speedDial(msg, 3))
		case "4":
			cmds = append(cmds, m.speedDial(msg, 4))
		case "5":
			cmds = append(cmds, m.speedDial(msg, 5))
		case "6":
			cmds = append(cmds, m.speedDial(msg, 6))
		case "7":
			cmds = append(cmds, m.speedDial(msg, 7))
		case "8":
			cmds = append(cmds, m.speedDial(msg, 8))
		case "9":
			cmds = append(cmds, m.speedDial(msg, 9))

		default:
			switch m.activePanel {
			case views.Instances:
				_, cmd = m.instances.Update(msg)
			case views.Configurations:
				_, cmd = m.configurations.Update(msg)
			case views.History:
				_, cmd = m.history.Update(msg)
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
		m.history.Update(m.historySize)
		m.statusBar.Update(m.statusSize)

	case configurations.ConfigurationSelectedMsg:
		m.selectedConfiguration = msg.Configuration

	case instances.InstanceSelectedMsg:
		m.selectedInstance = msg.Instance
		return m, tea.Quit

	case hist_view.ConnectionSelectedMsg:
		m.selectedHistoryConnection = msg.Connection
		return m, tea.Quit

	default:
		switch m.activePanel {
		case views.Instances:
			_, cmd = m.instances.Update(msg)
		case views.Configurations:
			_, cmd = m.configurations.Update(msg)
		case views.History:
			_, cmd = m.history.Update(msg)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	return lipgloss.JoinVertical(
		0, lipgloss.JoinHorizontal(
			0,
			views.BoxStyle(
				m.configSize, false).Render(m.configurations.View()),
			views.BoxStyle(
				m.instSize, false).Render(m.instances.View()),
		),
		views.BoxStyle(
			m.historySize, false).Render(m.history.View()),
		views.BoxStyle(
			m.statusSize, false).Render(m.statusBar.View()),
	)
}

func main() {
	for {
		p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
		r, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
		if m, ok := r.(*model); ok {
			if m.exited {
				fmt.Println("\n👋 See you soon!")
				os.Exit(0)
			}

			var selectedInstance *gcloud.Instance
			var selectedConfiguration string
			if m.selectedInstance != nil && m.selectedConfiguration != nil {
				selectedInstance = m.selectedInstance
				selectedConfiguration = m.selectedConfiguration.Name
			}
			if m.selectedHistoryConnection != nil {
				selectedInstance = m.selectedHistoryConnection.Instance
				selectedConfiguration = m.selectedHistoryConnection.ConfigName
			}

			if selectedInstance != nil {
				fmt.Println()
				fmt.Println(lipgloss.JoinHorizontal(
					0,
					lipgloss.NewStyle().Bold(true).Render("🚀 SSHing to instance "),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#7275ff")).Render(fmt.Sprintf("[%v]", selectedConfiguration)),
					lipgloss.NewStyle().Render(" -> "),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#ee6ff8")).Render(fmt.Sprintf("%v\n", selectedInstance.Name)),
					lipgloss.NewStyle().Render(" as "),
					lipgloss.NewStyle().Foreground(lipgloss.Color("#7275ff")).Render(config.Config.SSH.UserName),
					" ...",
				))
				fmt.Println()

				history.AddConnection(selectedConfiguration, selectedInstance)
				err = selectedInstance.SSH(selectedConfiguration)
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
