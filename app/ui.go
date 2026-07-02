package app

import (
	"fmt"
	"image/color"
	"runtime"
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
			PaddingBottom(1).
			AlignHorizontal(lipgloss.Center)

	labelStyle = lipgloss.NewStyle().Faint(true)

	textAreaStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Theme.Border)
)

// renderWaitMessage renders the idle state UI, displaying a prompt instructing
// the user to select a problem via the browser extension, alongside the compact
// help menu at the bottom of the screen.
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
	helpLayer := lipgloss.NewLayer(
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center,
			h.View(DefaultKeyMap)),
	).Y(m.height - 1)
	return lipgloss.NewCompositor(content, helpLayer).Render()
}

// renderInfo composes and renders the main active UI, combining the problem header,
// status metrics, and the split-pane test case viewports into a single layered view.
func (m Model) renderInfo() string {
	compositer := lipgloss.NewCompositor(m.renderHeader(), m.renderMiddle(), m.renderBody())
	return compositer.Render()
}

// renderHeader renders the top header layer containing the current problem's title,
// styled and constrained to the terminal width.
func (m Model) renderHeader() *lipgloss.Layer {
	content := headerStyle.Width(m.width).Render(m.Title)
	changedLayer := lipgloss.NewLayer("˙").X(m.width - 2)
	if m.fileChanged {
		return lipgloss.NewLayer(content, changedLayer)
	}
	return lipgloss.NewLayer(content)
}

// renderMiddle renders the middle status bar layer, displaying the current execution
// status (AC, WA, TLE, etc.), performance metrics (time/memory), and interactive
// test case navigation dots.
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
	var content string
	if m.status == Accepted || m.status == WrongAnswer {
		if runtime.GOOS == "windows" {
			content = fmt.Sprintf("Status: %s %.2fs", status, m.Tests[m.index].Time)
		} else {
			content = fmt.Sprintf("Status: %s %.2fs %.2fMiB", status, m.Tests[m.index].Time, float64(m.Tests[m.index].Memory)/1024)
		}
	} else {
		content = fmt.Sprintf("Status: %s", status)
	}
	statusLayer := lipgloss.NewLayer(content).X(1)

	dots := "  "
	for i := 0; i < len(m.Tests); i++ {
		var dot string
		if i == m.index {
			dot = "● "
		} else {
			dot = "○ "
		}
		var clr color.Color
		switch m.Tests[i].Status {
		case Accepted:
			clr = lipgloss.Green
		case NotAvailable:
			clr = lipgloss.White
		default:
			clr = lipgloss.Red
		}
		dots += lipgloss.NewStyle().Foreground(clr).Render(dot)
	}
	dotsLayers := lipgloss.NewLayer(dots).X(m.width - lipgloss.Width(dots))

	return lipgloss.NewLayer("", dotsLayers, statusLayer).Y(2)
}

// renderBody renders the main body layer containing the side-by-side split viewports.
// It dynamically labels the panes based on the current viewing mode (e.g., "Input" vs "Output").
func (m Model) renderBody() *lipgloss.Layer {
	h, w := m.height-5, m.width/2
	style := textAreaStyle.Height(h).Width(w)

	// select the appropriate labels
	labels := [][]string{
		{"Input", "Answer"},
		{"Input", "Output"},
		{"Input", "Error"},
		{"Answer", "Output"},
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

	return lipgloss.NewLayer("", leftLayer, rightLayer).Y(3)
}

// wrapContent formats a multi-line string for display in the UI viewports.
// It applies alternating background/foreground styles to adjacent lines to create
// a "zebra-stripe" effect for improved readability, and wraps the text to the specified width.
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
