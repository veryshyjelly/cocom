package app

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

var (
	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Fg)

	waitMessageStyle = lipgloss.NewStyle().
				Foreground(Theme.Fg)

	headerStyle = lipgloss.NewStyle().
			Padding(1, 0, 0).
			Foreground(Theme.Fg).
			AlignHorizontal(lipgloss.Center)

	textAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1).
			BorderForeground(Theme.Fg)
)

func (m Model) renderWaitMessage() string {
	s := "Select problem from competitive companion"
	style := waitMessageStyle.
		Padding((m.height-lipgloss.Height(s))/2, (m.width-lipgloss.Width(s))/2)
	return style.Render(s)
}

func (m Model) renderInfo() string {
	compositer := lipgloss.NewCompositor(m.renderHeader(), m.renderMiddle(), m.renderBody())
	return compositer.Render()
}

func (m Model) renderHeader() *lipgloss.Layer {
	content := headerStyle.Width(m.width).Render(m.Title)
	return lipgloss.NewLayer(content)
}

func (m Model) renderMiddle() *lipgloss.Layer {
	style := lipgloss.NewStyle()
	status := m.status
	if status == "NA" {
		status = "Status: " + style.Faint(true).Render(status)
	} else if status == "AC" {
		status = "Status: " + style.Foreground(Theme.Success).Render(status)
	} else {
		status = "Status: " + style.Foreground(Theme.Error).Render(status)
	}
	statusLayer := lipgloss.NewLayer(status).X(1)

	dots := "  "
	for i := 0; i < len(m.Tests); i++ {
		if i == m.index {
			dots += "● "
		} else {
			dots += "○ "
		}
	}
	dotsLayers := lipgloss.NewLayer(dots).X(m.width - lipgloss.Width(dots) - 2)

	return lipgloss.NewLayer("", dotsLayers, statusLayer).Y(2)
}

func (m Model) renderBody() *lipgloss.Layer {
	style := textAreaStyle.Height(m.height - 5).Width(m.width/2 - 1)
	headStyle := lipgloss.NewStyle().
		Faint(true).
		Width(style.GetWidth()).
		AlignHorizontal(lipgloss.Center)

	labels := []string{"", ""}
	switch m.mode {
	case InputAnswer:
		labels = []string{"Input", "Answer"}
	case InputOutput:
		labels = []string{"Input", "Output"}
	case AnswerOutput:
		labels = []string{"Answer", "Output"}
	case InputDiff:
		labels = []string{"Input", "Diff"}
	}

	// Create input layer
	leftLayer := lipgloss.NewLayer(
		style.Render(
			fmt.Sprintf("%s\n%s",
				headStyle.Render(labels[0]),
				m.leftViewPort.View(),
			),
		),
	)
	// Create output layer
	rightLayer := lipgloss.NewLayer(
		style.Render(
			fmt.Sprintf("%s\n%s",
				headStyle.Render(labels[1]),
				m.rightViewPort.View(),
			),
		),
	).
		X(m.width / 2)

	return lipgloss.NewLayer("", leftLayer, rightLayer).Y(3)
}
