package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gssh/views/configurations"
	"gssh/views/instances"
	"os"
)

var appStyle = lipgloss.NewStyle().Margin(1, 1)

type sessionState int

const (
	sessionStateInstances sessionState = iota
	sessionStateConfigurations
)

var _ tea.Model = model{}

type model struct {
	state          sessionState
	configurations tea.Model
	instances      tea.Model
}

func initialModel() model {
	return model{
		state:          sessionStateConfigurations,
		configurations: configurations.InitialModel(),
		instances:      instances.InitialModel(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.state = (m.state + 1) % 2
			if m.state == sessionStateConfigurations {
				m.configurations.Update(configurations.Activate)
				m.instances.Update(instances.Deactivate)
			}
			if m.state == sessionStateInstances {
				m.instances.Update(instances.Activate)
				m.configurations.Update(configurations.Deactivate)
			}
		default:
			switch m.state {
			case sessionStateInstances:
				_, newInstancesCmd := m.instances.Update(msg)
				cmd = newInstancesCmd

			case sessionStateConfigurations:
				_, newConfigurationsCmd := m.configurations.Update(msg)
				cmd = newConfigurationsCmd
			}

		}

	case tea.WindowSizeMsg:
		m.configurations.Update(msg)
		m.instances.Update(msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return appStyle.Render(lipgloss.JoinHorizontal(
		0,
		m.configurations.View(),
		m.instances.View(),
	))
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
