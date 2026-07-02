package app

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type KeyMap struct {
	Quit         key.Binding
	Run          key.Binding
	CreateFile   key.Binding
	CopyFile     key.Binding
	InputAnswer  key.Binding
	InputOutput  key.Binding
	InputError   key.Binding
	AnswerOutput key.Binding
	NextCase     key.Binding
	PreviousCase key.Binding
	Help         key.Binding
}

// ShortHelp returns a slice of key bindings to be displayed in the Bubble Tea
// help bubble's compact, single-line view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help, k.Quit,
	}
}

// FullHelp returns a 2D slice of key bindings organized into logical rows
// for the Bubble Tea help bubble's expanded, full-screen view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CreateFile, k.CopyFile, k.Run, k.NextCase, k.PreviousCase},
		{k.InputAnswer, k.InputOutput, k.InputError, k.AnswerOutput, k.Quit},
	}
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+d", "esc"),
		key.WithHelp("q/⎋", "quit")),
	Run: key.NewBinding(key.WithKeys("r", "ctrl+'"),
		key.WithHelp("r/⌃'", "run")),
	CreateFile: key.NewBinding(key.WithKeys("f"),
		key.WithHelp("f", "create file")),
	CopyFile: key.NewBinding(key.WithKeys("c", "ctrl+c"),
		key.WithHelp("c", "copy file")),
	InputAnswer: key.NewBinding(key.WithKeys("1", "j"),
		key.WithHelp("1/j", "input & answer")),
	InputOutput: key.NewBinding(key.WithKeys("2", "k"),
		key.WithHelp("2/k", "input & output")),
	InputError: key.NewBinding(key.WithKeys("3", "l"),
		key.WithHelp("3/l", "input & error")),
	AnswerOutput: key.NewBinding(key.WithKeys("4", ";"),
		key.WithHelp("4/;", "answer & output")),
	NextCase: key.NewBinding(key.WithKeys("tab", "right"),
		key.WithHelp("▶/⇥", "next case")),
	PreviousCase: key.NewBinding(key.WithKeys("shift+tab", "left"),
		key.WithHelp("◀/⇧⇥", "previous case")),
	Help: key.NewBinding(key.WithKeys("?"),
		key.WithHelp("?", "toggle help")),
}

// renderHelp renders the full-screen help interface displaying all available
// keyboard shortcuts. The view is dynamically centered within the current
// terminal dimensions.
func (m Model) renderHelp() string {
	h := help.New()
	h.ShowAll = true
	h.SetWidth(m.width)
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		h.View(DefaultKeyMap),
	)
}
