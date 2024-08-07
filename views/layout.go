package views

import (
	"github.com/charmbracelet/lipgloss"
	bl "github.com/winder/bubblelayout"
)

var PanelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)

func BoxStyle(size bl.Size, border bool) lipgloss.Style {
	width := size.Width
	height := size.Height
	if border {
		width -= 2
		height -= 2
	}
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
	if border {
		style = style.Border(lipgloss.RoundedBorder())
	}
	return style
}
