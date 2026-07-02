package app

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
)

type Model struct {
	Root string
	Config
	Problem
	Tests []Testcase

	status      Status
	fileChanged bool
	fileChan    chan string

	index         int
	height        int
	width         int
	ready         bool
	leftPane      Rect
	rightPane     Rect
	leftViewPort  viewport.Model
	rightViewPort viewport.Model

	mode Mode
}

type Mode int

const (
	InputAnswer Mode = iota
	InputOutput
	InputError
	AnswerOutput
	ShowHelp
)

// NewModel initializes and returns a new Bubble Tea Model with the provided
// project root directory and application configuration. It sets the initial
// execution status to NotAvailable.
func NewModel(root string, config Config, fileChan chan string) Model {
	log.Info("Initializing new model", "root", root)
	return Model{
		Root:     root,
		Config:   config,
		status:   NoData,
		fileChan: fileChan,
	}
}

// Init is the Bubble Tea initialization command. It returns nil as no initial
// asynchronous background tasks are required upon startup.
func (m Model) Init() tea.Cmd {
	log.Debug("Model Init called")
	return nil
}

// Update is the core Bubble Tea event loop. It processes keyboard inputs,
// window resizing, mouse events, and incoming problem/testcase messages to
// update the application state and trigger background commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		log.Debug("Key pressed", "key", msg.String())
		switch {
		case key.Matches(msg, DefaultKeyMap.Quit):
			log.Info("Quit command received")
			return m, tea.Quit
		case key.Matches(msg, DefaultKeyMap.Run) && len(m.Tests) > 0:
			log.Info("Run command triggered")
			m.status = Running
			return m, m.run
		case key.Matches(msg, DefaultKeyMap.CreateFile) && m.status != NoData:
			log.Info("Create file command triggered")
			return m, m.createFile
		case key.Matches(msg, DefaultKeyMap.CopyFile):
			log.Info("Copy file command triggered")
			m.fileChanged = false
			return m, m.copyFile
		case key.Matches(msg, DefaultKeyMap.InputAnswer):
			m.mode = InputAnswer
		case key.Matches(msg, DefaultKeyMap.InputOutput):
			m.mode = InputOutput
		case key.Matches(msg, DefaultKeyMap.InputError):
			m.mode = InputError
		case key.Matches(msg, DefaultKeyMap.AnswerOutput):
			m.mode = AnswerOutput
		case key.Matches(msg, DefaultKeyMap.NextCase) && len(m.Tests) > 0:
			m.index = (m.index + 1) % len(m.Tests)
			log.Debug("Next test case", "index", m.index)
		case key.Matches(msg, DefaultKeyMap.PreviousCase) && len(m.Tests) > 0:
			m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
			log.Debug("Previous test case", "index", m.index)
		case key.Matches(msg, DefaultKeyMap.Help):
			m.mode = ShowHelp * (1 - m.mode/ShowHelp)
		}
		m.updatePanes()
	case tea.MouseMsg:
		if m.rightPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.rightViewPort, cmd = m.rightViewPort.Update(msg)
		} else if m.leftPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.leftViewPort, cmd = m.leftViewPort.Update(msg)
		}
	case tea.WindowSizeMsg:
		log.Debug("Window resized", "width", msg.Width, "height", msg.Height)
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
	case Info:
		log.Info("Received new problem info", "title", msg.Name)
		m.setProblem(msg)
		if m.Config.CreateFile {
			log.Debug("Auto-creating file based on config")
			m.createFile()
		}
		m.fileChan <- m.getFileName()
		m.updatePanes()
	case []Testcase:
		log.Info("Received test case results", "count", len(msg))
		m.Tests = msg
		m.status = getFinalStatus(m.Tests)
		m.updatePanes()
		log.Info("Updated overall status", "status", m.status)
	case FileChanged:
		log.Info("Got filechange msg", "filename", msg)
		m.fileChanged = true
		if m.Config.RunOnSave && len(m.Tests) > 0 {
			m.status = Running
			return m, m.run
		}
	}
	return m, cmd
}

// View renders the current state of the application into a Bubble Tea View.
// It conditionally delegates rendering to the help screen, the idle waiting screen,
// or the main problem interface based on the current mode and URL state.
func (m Model) View() tea.View {
	var s string
	// conditionally render different views based on the mode
	switch {
	case m.mode == ShowHelp:
		s = m.renderHelp()
	case m.status == NoData:
		s = m.renderWaitMessage()
	default:
		s = m.renderInfo()
	}

	// render the view inside a bordered box
	s = containerStyle.
		Height(m.height + 2).
		Width(m.width + 2).
		Render(s)

	v := tea.NewView(s)
	v.MouseMode = tea.MouseModeCellMotion
	v.AltScreen = true
	return v
}

// updatePanes synchronizes the content of the left and right UI viewports with
// the currently selected test case and the active display mode (e.g., Input/Output,
// Input/Answer, Input/Error).
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
	}
}

// setLayout recalculates and applies the dimensions, padding, and coordinates
// for the split-pane UI layout based on the current terminal window size.
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
