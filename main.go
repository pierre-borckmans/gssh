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
	"os"
)

type sessionState int

const (
	sessionStateInstances sessionState = iota
	sessionStateConfigurations
)

var _ tea.Model = model{}

type model struct {
	layout                bl.BubbleLayout
	configurationsPanelId bl.ID
	instancesPanelId      bl.ID
	configSize            bl.Size
	instSize              bl.Size

	state          sessionState
	configurations tea.Model
	instances      tea.Model

	selectedConfiguration *gcloud.Configuration
	selectedInstance      *gcloud.Instance
}

func initialModel() model {
	layout := bl.New()
	configurationsPanelId := layout.Add("grow")
	instancesPanelId := layout.Add("grow")
	return model{
		layout:                layout,
		configurationsPanelId: configurationsPanelId,
		instancesPanelId:      instancesPanelId,
		state:                 sessionStateConfigurations,
		configurations:        configurations.InitialModel(),
		instances:             instances.InitialModel(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.configurations.Init(), m.instances.Init())
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
		case "escape", "ctrl+c":
			return m, tea.Quit
		case "left", "shift+tab":
			m.state = (m.state - 1 + 2) % 2
			if m.state == sessionStateConfigurations {
				m.configurations.Update(configurations.FocusMsg{})
				m.instances.Update(instances.BlurMsg{})
			}
			if m.state == sessionStateInstances {
				m.instances.Update(instances.FocusMsg{})
				m.configurations.Update(configurations.BlurMsg{})
			}
		case "right", "tab":
			m.state = (m.state + 1) % 2
			if m.state == sessionStateConfigurations {
				m.configurations.Update(configurations.FocusMsg{})
				m.instances.Update(instances.BlurMsg{})
			}
			if m.state == sessionStateInstances {
				m.instances.Update(instances.FocusMsg{})
				m.configurations.Update(configurations.BlurMsg{})
			}
		default:
			switch m.state {
			case sessionStateInstances:
				_, cmd = m.instances.Update(msg)
			case sessionStateConfigurations:
				_, cmd = m.configurations.Update(msg)
			}
		}

	case tea.WindowSizeMsg:
		return m, func() tea.Msg {
			return m.layout.Resize(msg.Width, msg.Height)
		}

	case bl.BubbleLayoutMsg:
		m.configSize, _ = msg.Size(m.configurationsPanelId)
		m.instSize, _ = msg.Size(m.instancesPanelId)
		m.configurations.Update(m.configSize)
		m.instances.Update(m.instSize)

	case configurations.ConfigurationSelectedMsg:
		m.selectedConfiguration = msg.Configuration

	case instances.InstanceSelectedMsg:
		m.selectedInstance = msg.Instance
		return m, tea.Quit

	default:
		switch m.state {
		case sessionStateInstances:
			_, cmd = m.instances.Update(msg)
		case sessionStateConfigurations:
			_, cmd = m.configurations.Update(msg)
		}
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinHorizontal(
		0,
		views.BoxStyle(
			m.configSize, false).Render(m.configurations.View()),
		views.BoxStyle(
			m.instSize, false).Render(m.instances.View()),
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	r, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	if m, ok := r.(model); ok {
		if m.selectedInstance == nil || m.selectedConfiguration == nil {
			os.Exit(0)
		}

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
	}
}
