package app

import (
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Root string
	Config
	Problem
	Tests []Testcase

	status string
	index  int
	height int
	width  int

	leftViewPort  viewport.Model
	rightViewPort viewport.Model
	ready         bool

	leftPane  Rect
	rightPane Rect

	mode Mode
}

type Mode int

const (
	InputAnswer Mode = iota
	InputOutput
	AnswerOutput
	InputDiff
)

func NewModel(root string, config Config) Model {
	return Model{
		Root:   root,
		Config: config,
		status: "NA",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		/*
			Functionality
		*/
		case "r":
			return m, m.run
		case "f":
			return m, m.createFile
		/*
			Handle Navigation
		*/
		case "space":
			if m.mode == InputAnswer {
				m.mode = InputOutput
			} else {
				m.mode = InputAnswer
			}
		case "o":
			if m.mode != AnswerOutput {
				m.mode = AnswerOutput
			} else {
				m.mode = InputAnswer
			}
		case "d":
			if m.mode != InputDiff {
				m.mode = InputDiff
			} else {
				m.mode = InputAnswer
			}
		case "tab":
			m.index = (m.index + 1) % len(m.Tests)
		case "shift+tab":
			m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
		}
		m.updatePanes()
	case Info:
		m.setProblem(msg)
		m.updatePanes()
	case tea.WindowSizeMsg:
		if !m.ready {
			m.leftViewPort = viewport.New()
			m.leftViewPort.YPosition = 4
			m.rightViewPort = viewport.New()
			m.rightViewPort.YPosition = 4
			m.ready = true
		}
		m.width = msg.Width - 2
		m.height = msg.Height - 2
		m.setLayout()
		m.updatePanes()
	case tea.MouseMsg:
		if m.rightPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.rightViewPort, cmd = m.rightViewPort.Update(msg)
		} else if m.leftPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.leftViewPort, cmd = m.leftViewPort.Update(msg)
		}
	}

	return m, cmd
}

func (m Model) View() tea.View {
	var s string

	if m.Url == "" {
		s = m.renderWaitMessage()
	} else {
		s = m.renderInfo()
	}

	s = containerStyle.
		Height(m.height + 2).
		Width(m.width + 2).
		Render(s)

	v := tea.NewView(s)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}

// updatePanes updates the content of the left and right viewports based on the current display mode.
// It retrieves the test case specified by the model's current index.
// The content displayed in each viewport depends on the value of `m.mode`.
// For InputOutput mode, it shows the test case's Input and Output.
// For InputAnswer mode, it shows the test case's Input and Answer.
// For AnswerOutput mode, it shows the test case's Answer and Output.
// For InputDiff mode, it displays the test case's Input and Answer.
func (m *Model) updatePanes() {
	if len(m.Tests) == 0 {
		return
	}

	width := m.leftPane.W

	testCase := m.Tests[m.index]
	input, output, answer := wrapContent(testCase.Input, width),
		wrapContent(testCase.Output, width), wrapContent(testCase.Answer, width)
	if m.mode == InputOutput {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(output)
	} else if m.mode == InputAnswer {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(answer)
	} else if m.mode == AnswerOutput {
		m.leftViewPort.SetContent(answer)
		m.rightViewPort.SetContent(output)
	} else if m.mode == InputDiff {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(answer)
	}
}

// setLayout calculates and sets the dimensions and positions of the left and right display panes.
// // It updates the size of the corresponding viewports based on the model's current width and height.
func (m *Model) setLayout() {
	m.leftPane = Rect{
		X: 0,
		Y: 4,
		// 2 for padding and 2 for border
		W: m.width/2 - 4,
		H: m.height - 4,
	}
	// 1 for label
	m.leftViewPort.SetHeight(m.leftPane.H - 1)
	m.leftViewPort.SetWidth(m.leftPane.W)

	m.rightPane = Rect{
		X: (m.width + 1) / 2,
		Y: 4,
		// 2 for padding and 2 for border
		W: m.width/2 - 4,
		H: m.height - 4,
	}
	// 1 for label
	m.rightViewPort.SetHeight(m.rightPane.H - 1)
	m.rightViewPort.SetWidth(m.rightPane.W)
}
