package app

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type ColorScheme struct {
	Bg        color.Color
	Fg        color.Color
	Primary   color.Color
	Secondary color.Color
	Accent    color.Color
	Success   color.Color
	Warning   color.Color
	Error     color.Color
}

var (
	Gold = ColorScheme{
		Fg:        lipgloss.Color("#FAE7CB"),
		Primary:   lipgloss.Color("#59B292"),
		Secondary: lipgloss.Color("#FFC94D"),
		Success:   lipgloss.Green,
		Warning:   lipgloss.BrightCyan,
		Error:     lipgloss.Color("#FA6781"),
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
