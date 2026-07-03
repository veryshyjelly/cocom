package server

import (
	"encoding/json"
	"net/http"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/veryshyjelly/cocom/core"
)

// HandleData returns an HTTP handler function that listens for incoming JSON payloads
// from the Competitive Companion browser extension. It parses the problem data and
// injects it into the Bubble Tea event loop via the provided Program instance.
func HandleData(p *tea.Program) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Received HTTP request", "method", r.Method, "remote", r.RemoteAddr)
		var data core.Info
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
