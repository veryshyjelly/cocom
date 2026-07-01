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
	InputAnswer  key.Binding
	InputOutput  key.Binding
	InputError   key.Binding
	InputDiff    key.Binding
	AnswerOutput key.Binding
	NextCase     key.Binding
	PreviousCase key.Binding
	Help         key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Help, k.Quit,
	}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Help, k.Quit, k.Run, k.NextCase, k.PreviousCase},
		{k.InputAnswer, k.InputOutput, k.InputError, k.InputDiff, k.AnswerOutput},
	}
}

var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q/⎋", "quit")),
	Run: key.NewBinding(key.WithKeys("r", "ctrl+'"),
		key.WithHelp("r/⌃'", "run")),
	CreateFile: key.NewBinding(key.WithKeys("f"),
		key.WithHelp("f", "create file")),
	InputAnswer: key.NewBinding(key.WithKeys("1", "i"),
		key.WithHelp("1/i", "input & answer")),
	InputOutput: key.NewBinding(key.WithKeys("2", "o"),
		key.WithHelp("2/o", "input & output")),
	InputError: key.NewBinding(key.WithKeys("3", "e"),
		key.WithHelp("3/e", "input & error")),
	InputDiff: key.NewBinding(key.WithKeys("4", "d"),
		key.WithHelp("4/d", "input & diff")),
	AnswerOutput: key.NewBinding(key.WithKeys("5", "c"),
		key.WithHelp("5/c", "answer & output")),
	NextCase: key.NewBinding(key.WithKeys("tab", "right"),
		key.WithHelp("▶/⇥", "next case")),
	PreviousCase: key.NewBinding(key.WithKeys("shift+tab", "left"),
		key.WithHelp("◀/⇧⇥", "previous case")),
	Help: key.NewBinding(key.WithKeys("?"),
		key.WithHelp("?", "help")),
}

// renderHelp renders the help view for the model using the default key map.
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
