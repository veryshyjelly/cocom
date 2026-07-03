package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ColorScheme struct {
	Foreground color.Color
	Border     color.Color
	Success    color.Color
	Warning    color.Color
	Error      color.Color
}

var (
	Gold = ColorScheme{
		Foreground: lipgloss.White,
		Border:     lipgloss.Color("#FAE7CB"),
		Success:    lipgloss.Green,
		Warning:    lipgloss.BrightCyan,
		Error:      lipgloss.Color("#FA6781"),
	}
)

var Theme = Gold

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

type Rect struct {
	X, Y int
	W, H int
}

// Contains checks if a given 2D coordinate (x, y) falls within the boundaries
// of the rectangular UI pane. Used for routing mouse events to the correct viewport.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X &&
		x < r.X+r.W &&
		y >= r.Y &&
		y < r.Y+r.H
}
