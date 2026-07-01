package app

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/bubbles/v2/help"
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
	content := lipgloss.NewLayer(lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		waitMessageStyle.Render(s),
	))
	h := help.New()
	h.SetWidth(m.width)
	helpLayer := lipgloss.NewLayer(
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center,
			h.View(DefaultKeyMap)),
	).Y(m.height - 1)
	return lipgloss.NewCompositor(content, helpLayer).Render()
}

func (m Model) renderInfo() string {
	compositer := lipgloss.NewCompositor(m.renderHeader(), m.renderMiddle(), m.renderBody())
	return compositer.Render()
}

func (m Model) renderHeader() *lipgloss.Layer {
	content := headerStyle.Width(m.width).Render(m.Title)
	//message := labelStyle.Render("? toggle help")
	//helpLayer := lipgloss.NewLayer(message).X(m.width - lipgloss.Width(message) - 1)
	return lipgloss.NewLayer(content)
}

func (m Model) renderMiddle() *lipgloss.Layer {
	style := lipgloss.NewStyle()
	status := string(m.status)
	switch m.status {
	case NotAvailable:
		status = style.Faint(true).Render(status)
	case Running:
		status = style.Faint(true).Foreground(Theme.Warning).Render(status)
	case Accepted:
		status = style.Foreground(Theme.Success).Render(status)
	default:
		status = style.Foreground(Theme.Error).Render(status)
	}
	statusLayer := lipgloss.NewLayer("Status: " + status).X(1)

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
	dotsLayers := lipgloss.NewLayer(dots).X(m.width - lipgloss.Width(dots))

	return lipgloss.NewLayer("", dotsLayers, statusLayer).Y(1)
}

func (m Model) renderBody() *lipgloss.Layer {
	h, w := m.height-4, m.width/2
	style := textAreaStyle.Height(h).Width(w)

	// select the appropriate labels
	labels := [][]string{
		{"Input", "Answer"},
		{"Input", "Output"},
		{"Input", "Error"},
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

func wrapContent(content string, width int) string {
	oddStyle := lipgloss.NewStyle()
	evenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E6F0FF"))

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i%2 == 0 {
			lines[i] = oddStyle.Render(line)
		} else {
			lines[i] = evenStyle.Render(line)
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Render(strings.Join(lines, "\n"))
}
