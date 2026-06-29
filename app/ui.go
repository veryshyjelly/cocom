package app

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Magenta)

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Cyan).
			Padding(0, 2)
	headerStyle = lipgloss.NewStyle()
)

func (m Model) renderWaitMessage() string {
	return fmt.Sprintf("\n %s %s \n", m.spinner.View(), "Waiting for data...")
}

func (m Model) renderInfo() string {
	s := fmt.Sprintf("%s\n%s\n%s",
		m.renderHeader(), m.renderMiddle(), m.renderBody())
	return s
}

func (m Model) renderHeader() string {
	return ""
}

func (m Model) renderMiddle() string {
	return ""
}

func (m Model) renderBody() string {
	return ""
}
