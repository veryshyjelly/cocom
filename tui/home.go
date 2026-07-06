package tui

import (
	"fmt"
	"image/color"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/veryshyjelly/cocom/core"
)

// View renders the current state of the application into a Bubble Tea View.
// It conditionally delegates rendering to the help screen, the idle waiting screen,
// or the main problem interface based on the current mode and URL state.
func (m Model) View() tea.View {
	// compose all the layers
	compositer := lipgloss.NewCompositor(m.renderHeader(), m.renderMiddle(), m.renderBody())
	// render the view inside a bordered box
	s := containerStyle.
		Height(m.height + 2).
		Width(m.width + 2).
		Render(compositer.Render())

	v := tea.NewView(s)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
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
	status := string(m.Status)
	switch m.Status {
	case core.NotAvailable:
		status = style.Faint(true).Render(status)
	case core.Running:
		status = style.Faint(true).Foreground(Theme.Warning).Render(status)
	case core.Accepted:
		status = style.Foreground(Theme.Success).Render(status)
	default:
		status = style.Foreground(Theme.Error).Render(status)
	}
	var content string
	if m.Status == core.Accepted || m.Status == core.WrongAnswer {
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
		case core.Accepted:
			clr = lipgloss.Green
		case core.NotAvailable:
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
	var h, w int
	var leftX, leftY int
	var rightX, rightY int

	if m.orientation == Vertical {
		h = (m.height - 5) / 2
		w = m.width
		leftX = 0
		leftY = 0
		rightX = 0
		rightY = (m.height-5)/2 + 1
	} else {
		h = m.height - 5
		w = m.width / 2
		leftX = 0
		leftY = 0
		rightX = (m.width + 1) / 2
		rightY = 0
	}

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
	).X(leftX).Y(leftY)

	// Create output layer
	rightLayer := lipgloss.NewLayer(
		style.Render(
			fmt.Sprintf("%s\n%s",
				lipgloss.PlaceHorizontal(w-1, lipgloss.Center, labelStyle.Render(labels[1])),
				m.rightViewPort.View(),
			),
		),
	).X(rightX).Y(rightY)

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
