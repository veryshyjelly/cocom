package app

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border)

	waitMessageStyle = lipgloss.NewStyle().
				Foreground(Theme.Foreground)

	headerStyle = lipgloss.NewStyle().
			Foreground(Theme.Foreground).
			AlignHorizontal(lipgloss.Center)

	labelStyle = lipgloss.NewStyle().Faint(true)

	textAreaStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border)
)

func (m Model) renderWaitMessage() string {
	s := "Select problem from competitive companion"
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		waitMessageStyle.Render(s),
	)
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
		var dot string
		if i == m.index {
			dot = "● "
		} else {
			dot = "○ "
		}
		var clr color.Color
		if m.Tests[i].Status == "AC" {
			clr = lipgloss.Green
		} else if m.Tests[i].Status == "" {
			clr = lipgloss.White
		} else {
			clr = lipgloss.Red
		}
		dots += lipgloss.NewStyle().Foreground(clr).Render(dot)
	}
	dotsLayers := lipgloss.NewLayer(dots).X(m.width - lipgloss.Width(dots) - 2)

	return lipgloss.NewLayer("", dotsLayers, statusLayer).Y(1)
}

func (m Model) renderBody() *lipgloss.Layer {
	h, w := m.height-4, m.width/2
	style := textAreaStyle.Height(h).Width(w)

	// select the appropriate labels
	labels := [][]string{
		{"Input", "Answer"},
		{"Input", "Output"},
		{"Answer", "Output"},
		{"Input", "Diff"},
	}[m.mode]

	// Create input layer
	leftLayer := lipgloss.NewLayer(
		style.Render(
			fmt.Sprintf("%s\n%s",
				lipgloss.PlaceHorizontal(w-1, lipgloss.Center, labelStyle.Render(labels[0])),
				m.leftViewPort.View(),
			),
		),
	)
	// Create output layer
	rightLayer := lipgloss.NewLayer(
		style.Render(
			fmt.Sprintf("%s\n%s",
				lipgloss.PlaceHorizontal(w-1, lipgloss.Center, labelStyle.Render(labels[1])),
				m.rightViewPort.View(),
			),
		),
	).
		X((m.width + 1) / 2)

	return lipgloss.NewLayer("", leftLayer, rightLayer).Y(2)
}
