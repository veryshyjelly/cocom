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
	MemoryLimit uint64 `json:"memoryLimit"`
	TimeLimit   uint64 `json:"timeLimit"`
	Tests       []Test `json:"tests"`
}

type Test struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// HandleData returns an HTTP handler function that listens for incoming JSON payloads
// from the Competitive Companion browser extension. It parses the problem data and
// injects it into the Bubble Tea event loop via the provided Program instance.
func HandleData(p *tea.Program) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Received HTTP request", "method", r.Method, "remote", r.RemoteAddr)
		var data Info
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Error("Failed to parse incoming JSON data", "err", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		log.Debug("Parsed problem data", "title", data.Name, "url", data.Url, "tests_count", len(data.Tests))
		p.Send(data)
		w.WriteHeader(http.StatusOK)
	}
}

// setProblem updates the model's internal state with a new competitive programming problem.
// It initializes the test cases, resets the current test index to prevent out-of-bounds errors,
// and resets the overall execution status.
func (m *Model) setProblem(info Info) {
	log.Info("Setting new problem in model", "title", info.Name, "url", info.Url)
	m.status = NotAvailable
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
			Status: NotAvailable,
		})
	}
}
