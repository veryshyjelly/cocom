package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/log/v2"
	"github.com/veryshyjelly/cocom/core"
)

type AddTestCase struct {
	parent Model
	form   huh.Model

	input  string
	output string
}

func NewAddTestCase(parent Model) *AddTestCase {
	t := &AddTestCase{parent: parent}

	t.form = huh.NewForm(
		huh.NewGroup(
			huh.NewText().Title("Stdin").Value(&t.input),
			huh.NewText().Title("Stdout").Value(&t.output),
		),
	)

	return t
}

func (m *AddTestCase) Init() tea.Cmd {
	return m.form.Init()
}

func (m *AddTestCase) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)

	if m.form.(*huh.Form).State == huh.StateCompleted {
		log.Debug("Adding test case", "input", m.input, "output", m.output)
		m.parent.Tests = append(m.parent.Tests,
			core.Testcase{
				Input:  m.input,
				Answer: m.output,
				Status: core.NotAvailable,
			},
		)
		return m.parent, nil
	}

	return m, cmd
}

func (m *AddTestCase) View() tea.View {
	v := tea.NewView(m.form.View())
	v.AltScreen = true
	return v
}
