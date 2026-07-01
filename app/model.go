package app

import (
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Root string
	Config
	Problem
	Tests []Testcase

	status Status
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
	InputError
	AnswerOutput
	InputDiff
	ShowHelp
)

func NewModel(root string, config Config) Model {
	return Model{
		Root:   root,
		Config: config,
		status: NotAvailable,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model, returning the updated model and a command.
// It processes key presses, window size changes, mouse events, and test case results.
// The method updates the model's state, status, viewports, and active test case index.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Run) && len(m.Tests) > 0:
			m.status = Running
			return m, m.run
		case key.Matches(msg, DefaultKeyMap.CreateFile):
			return m, m.createFile
		case key.Matches(msg, DefaultKeyMap.InputAnswer):
			m.mode = InputAnswer
		case key.Matches(msg, DefaultKeyMap.InputOutput):
			m.mode = InputOutput
		case key.Matches(msg, DefaultKeyMap.InputError):
			m.mode = InputError
		case key.Matches(msg, DefaultKeyMap.InputDiff):
			m.mode = InputDiff
		case key.Matches(msg, DefaultKeyMap.AnswerOutput):
			m.mode = AnswerOutput
		case key.Matches(msg, DefaultKeyMap.NextCase) && len(m.Tests) > 0:
			m.index = (m.index + 1) % len(m.Tests)
		case key.Matches(msg, DefaultKeyMap.PreviousCase) && len(m.Tests) > 0:
			m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
		case key.Matches(msg, DefaultKeyMap.Help):
			m.mode = ShowHelp - m.mode/ShowHelp
		}
		m.updatePanes()
	case Info:
		m.status = NotAvailable
		m.setProblem(msg)
		if m.Config.CreateFile {
			m.createFile()
		}
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
	case []Testcase:
		m.Tests = msg
		switch {
		case slices.ContainsFunc(m.Tests,
			func(t Testcase) bool { return t.Status == CompilationError }):
			m.status = CompilationError
		case slices.ContainsFunc(m.Tests,
			func(t Testcase) bool { return t.Status == RuntimeError }):
			m.status = RuntimeError
		case slices.ContainsFunc(m.Tests,
			func(t Testcase) bool { return t.Status == WrongAnswer }):
			m.status = WrongAnswer
		default:
			m.status = Accepted
		}
	}

	return m, cmd
}

// View renders the current state of the Model into a tea.View for display.
// It conditionally renders help, a wait message, or information based on the model's state.
// The rendered content is wrapped within a styled container.
// The returned view is configured for cell motion mouse mode and uses the alternate screen buffer.
func (m Model) View() tea.View {
	var s string

	if m.mode == ShowHelp {
		s = m.renderHelp()
	} else if m.Url == "" {
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
	erro := wrapContent(testCase.Error, width)
	if m.mode == InputOutput {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(output)
	} else if m.mode == InputAnswer {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(answer)
	} else if m.mode == InputError {
		m.leftViewPort.SetContent(input)
		m.rightViewPort.SetContent(erro)
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
		Y: 5,
		// 2 for padding and 2 for border
		W: m.width/2 - 4,
		// 2 for border 2 for heading and 1 for status
		H: m.height - 5,
	}
	// 1 for label
	m.leftViewPort.SetHeight(m.leftPane.H - 1)
	m.leftViewPort.SetWidth(m.leftPane.W)

	m.rightPane = Rect{
		X: (m.width + 1) / 2,
		Y: 5,
		// 2 for padding and 2 for border
		W: m.width/2 - 4,
		// 2 for border 2 for heading and 1 for status
		H: m.height - 5,
	}
	// 1 for label
	m.rightViewPort.SetHeight(m.rightPane.H - 1)
	m.rightViewPort.SetWidth(m.rightPane.W)
}
