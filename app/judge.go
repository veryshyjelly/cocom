package app

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/google/shlex"
)

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
	Status Status
	Time   float64
}

type Status string

const (
	NotAvailable     Status = "NA"
	Accepted         Status = "AC"
	RuntimeError     Status = "RE"
	CompilationError Status = "CE"
	WrongAnswer      Status = "WA"
	Running          Status = "WIP"
)

func (m Model) compile() (string, error) {
	log.Debug("creating sandbox directory")
	dir, err := os.MkdirTemp("", "sandbox")
	unwrap("unable to create temporary directory", err)

	solutionFile := m.getSolution()
	filePath := filepath.Join(dir, m.Compiler.Source)
	err = os.WriteFile(filePath, []byte(solutionFile), 0644)
	unwrap("unable to write solution file", err)

	var stderr bytes.Buffer
	cmd := exec.Command(m.Compile, m.Compiler.Args...)
	cmd.Dir = dir
	cmd.Stderr = &stderr
	err = cmd.Run()

	logger.Infof("compilation command %#v", cmd)

	if err != nil {
		logger.Error("compilation failed", stderr.String())
		return "", errors.New(stderr.String())
	}

	return dir, nil
}

func (m Model) run() tea.Msg {
	tests := m.Tests

	dir, err := m.compile()
	defer os.RemoveAll(dir)
	if err != nil {
		for i := range tests {
			tests[i].Status = CompilationError
			tests[i].Error = err.Error()
		}
		return tests
	}

	args, err := shlex.Split(m.Run)
	if err != nil {
		unwrap("unable to parse run arguments", err)
	}

	for i := range tests {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Stdin = strings.NewReader(tests[i].Input)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		tests[i].Output = stdout.String()
		tests[i].Error = stderr.String()
		if err != nil {
			tests[i].Status = RuntimeError
		} else if strings.TrimSpace(tests[i].Output) == strings.TrimSpace(tests[i].Answer) {
			tests[i].Status = Accepted
		} else {
			tests[i].Status = WrongAnswer
		}
	}

	return tests
}
