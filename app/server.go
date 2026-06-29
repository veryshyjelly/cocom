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
	Tests       []Test `json:"test"`
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

func setProblem(info Info, m Model) Model {
	// Fill problem and test case in model
	m.Problem = Problem{
		Title:       info.Name,
		Url:         info.Url,
		MemoryLimit: info.MemoryLimit,
		TimeLimit:   info.TimeLimit,
	}

	m.Tests = make([]Testcase, len(m.Tests))
	for i, t := range info.Tests {
		m.Tests[i] = Testcase{
			Input:  t.Input,
			Output: t.Output,
		}
	}

	return m
}
