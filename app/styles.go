package app

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

type Rect struct {
	X, Y int
	W, H int
}

func (r Rect) Contains(x, y int) bool {
	return x >= r.X &&
		x < r.X+r.W &&
		y >= r.Y &&
		y < r.Y+r.H
}
