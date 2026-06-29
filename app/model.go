package app

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Root string
	Config
	Problem
	Tests   []Testcase
	height  int
	width   int
	spinner spinner.Model
}

func NewModel(root string, config Config) Model {
	spin := spinner.New(spinner.WithSpinner(spinner.Pulse))
	spin.Style = spinnerStyle

	return Model{
		Root:    root,
		Config:  config,
		spinner: spin,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case Info:
		m = setProblem(msg, m)
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, cmd
}

func (m Model) View() tea.View {
	var s string

	if m.Url == "" {
		s += m.renderWaitMessage()
	} else {
		s += m.renderInfo()
	}

	s = containerStyle.
		Height(m.height - 2).
		Width(m.width - 2).
		Render(s)

	return tea.NewView(s)
}
