package app

import (
	"encoding/json"
	"net/http"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
)

type Info struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Url         string `json:"url"`
	Interactive bool   `json:"interactive"`
	MemoryLimit uint64 `json:"memory_limit"`
	TimeLimit   uint64 `json:"time_limit"`
	Tests       []Test `json:"tests"`
}

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

func HandleData(p *tea.Program) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data Info

		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Error("cannot parse data", "err", err)
			return
		}

		p.Send(data)
	}
}

func (m *Model) setProblem(info Info) {
	m.index = min(m.index, len(info.Tests)-1)

	// Fill problem and test case in model
	m.Problem = Problem{
		Title:       info.Name,
		Url:         info.Url,
		MemoryLimit: info.MemoryLimit,
		TimeLimit:   info.TimeLimit,
	}

	m.Tests = make([]Testcase, 0, len(info.Tests))
	for _, t := range info.Tests {
		m.Tests = append(m.Tests, Testcase{
			Input:  t.Input,
			Answer: t.Output,
		})
	}
}
