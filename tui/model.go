package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/veryshyjelly/cocom/config"
	"github.com/veryshyjelly/cocom/core"
)

type Model struct {
	core.App

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

	mode        Mode
	orientation Orientation
}

type Mode int

const (
	InputAnswer Mode = iota
	InputOutput
	InputError
	AnswerOutput
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// NewModel initializes and returns a new Bubble Tea Model with the provided
// project root directory and application configuration. It sets the initial
// execution status to NotAvailable.
func NewModel(root string, config config.Config, fileChan chan string) Model {
	log.Info("Initializing new model", "root", root)
	return Model{
		App:         core.App{Root: root, Config: config},
		fileChan:    fileChan,
		orientation: Horizontal,
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
		case key.Matches(msg, DefaultKeyMap.Run):
			log.Info("Run command triggered")
			m.Status = core.Running
			return m, m.Run
		case key.Matches(msg, DefaultKeyMap.CreateFile):
			log.Info("Create file command triggered")
			m.CreateFile()
			if m.Config.Editor != "" {
				return m, m.OpenEditor()
			}
		case key.Matches(msg, DefaultKeyMap.CopyFile):
			log.Info("Copy file command triggered")
			m.fileChanged = false
			return m, m.CopyFile
		case key.Matches(msg, DefaultKeyMap.AddCase):
			form := NewAddTestCase(m)
			return form, form.Init()
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
			hi := Help{parent: m, height: m.height, width: m.width}
			return hi, nil
		case key.Matches(msg, DefaultKeyMap.HorizontalLayout):
			m.orientation = Horizontal
			m = m.setLayout()
		case key.Matches(msg, DefaultKeyMap.VerticalLayout):
			m.orientation = Vertical
			m = m.setLayout()
		}
	case tea.MouseMsg:
		if m.rightPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.rightViewPort, cmd = m.rightViewPort.Update(msg)
		} else if m.leftPane.Contains(msg.Mouse().X, msg.Mouse().Y) {
			m.leftViewPort, cmd = m.leftViewPort.Update(msg)
		} else if len(m.Tests) > 0 {
			switch msg.Mouse().Button {
			case tea.MouseWheelUp:
				m.index = (m.index + 1) % len(m.Tests)
			case tea.MouseWheelDown:
				m.index = (m.index - 1 + len(m.Tests)) % len(m.Tests)
			}
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
		m = m.setLayout()
	case core.Info:
		log.Info("Received new problem info", "title", msg.Name)
		m = m.setProblem(msg)
		if m.Config.CreateFile {
			log.Debug("Auto-creating file based on config")
			m.CreateFile()
		}
		m.fileChan <- m.GetFileName()
		if m.Config.Editor != "" {
			return m, m.OpenEditor()
		}
	case []core.Testcase:
		log.Info("Received test case results", "count", len(msg))
		m.Tests = msg
		m.Status = core.GetFinalStatus(m.Tests)
		log.Info("Updated overall status", "status", m.Status)
	case fsnotify.Event:
		log.Info("Got filechange msg", "filename", msg)
		m.fileChanged = true
		if m.Config.RunOnSave {
			m.Status = core.Running
			return m, m.Run
		}
	}
	return m.updatePanes(), cmd
}

// updatePanes synchronizes the content of the left and right UI viewports with
// the currently selected test case and the active display mode (e.g., Input/Output,
// Input/Answer, Input/Error).
func (m Model) updatePanes() Model {
	if len(m.Tests) == 0 {
		return m
	}

	width := m.leftPane.W

	testCase := m.Tests[m.index]
	input, output, answer := wrapContent(testCase.Input, width),
		wrapContent(testCase.Output, width), wrapContent(testCase.Answer, width)
	erro := wrapContent(testCase.Error, width)

	m.leftViewPort.SetContent(input)
	switch m.mode {
	case InputOutput:
		m.rightViewPort.SetContent(output)
	case InputAnswer:
		m.rightViewPort.SetContent(answer)
	case InputError:
		m.rightViewPort.SetContent(erro)
	case AnswerOutput:
		m.leftViewPort.SetContent(answer)
		m.rightViewPort.SetContent(output)
	}

	return m
}

// setLayout recalculates and applies the dimensions, padding, and coordinates
// for the split-pane UI layout based on the current terminal window size.
func (m Model) setLayout() Model {
	if m.orientation == Vertical {
		m.leftPane = Rect{
			X: 0,
			Y: 5,
			W: m.width - 4,
			H: (m.height - 5) / 2,
		}
		m.leftViewPort.SetHeight(m.leftPane.H - 1)
		m.leftViewPort.SetWidth(m.leftPane.W)

		m.rightPane = Rect{
			X: 0,
			Y: 5 + (m.height-5)/2 + 1,
			W: m.width - 4,
			H: (m.height - 5) / 2,
		}
		m.rightViewPort.SetHeight(m.rightPane.H - 1)
		m.rightViewPort.SetWidth(m.rightPane.W)
	} else {
		m.leftPane = Rect{
			X: 0,
			Y: 5,
			W: m.width/2 - 4,
			H: m.height - 5,
		}
		m.leftViewPort.SetHeight(m.leftPane.H - 1)
		m.leftViewPort.SetWidth(m.leftPane.W)

		m.rightPane = Rect{
			X: (m.width + 1) / 2,
			Y: 5,
			W: m.width/2 - 4,
			H: m.height - 5,
		}
		m.rightViewPort.SetHeight(m.rightPane.H - 1)
		m.rightViewPort.SetWidth(m.rightPane.W)
	}
	return m
}

// setProblem updates the model's internal state with a new competitive programming problem.
// It initializes the test cases, resets the current test index to prevent out-of-bounds errors,
// and resets the overall execution status.
func (m Model) setProblem(info core.Info) Model {
	log.Info("Setting new problem in model", "title", info.Name, "url", info.Url)
	m.Status = core.NotAvailable
	m.index = min(m.index, len(info.Tests)-1)
	// Fill problem and test case in model
	m.Problem = core.Problem{
		Title:       info.Name,
		Url:         info.Url,
		MemoryLimit: info.MemoryLimit,
		TimeLimit:   info.TimeLimit,
	}
	m.Tests = make([]core.Testcase, 0, len(info.Tests))
	for _, t := range info.Tests {
		m.Tests = append(m.Tests, core.Testcase{
			Input:  t.Input,
			Answer: t.Output,
			Status: core.NotAvailable,
		})
	}
	return m
}
