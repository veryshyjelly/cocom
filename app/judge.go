package app

import tea "charm.land/bubbletea/v2"

type Problem struct {
	Title       string
	Url         string
	MemoryLimit uint64
	TimeLimit   uint64
}

type Testcase struct {
	Input  string
	Output string
	Error  string
	Answer string
	Status string
	Time   float64
}

func (m Model) Compile() error {
	return nil
}

func (m Model) Run() tea.Msg {
	return nil
}
