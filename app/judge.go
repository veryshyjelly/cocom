package app

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/google/shlex"
	"github.com/veryshyjelly/cocom/app/memory"
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
	Memory uint64 // in kbs
}

type Status string

const (
	NoData           Status = "ND"
	NotAvailable     Status = "NA"
	Accepted         Status = "AC"
	RuntimeError     Status = "RE"
	CompilationError Status = "CE"
	TimeLimitError   Status = "TLE"
	WrongAnswer      Status = "WA"
	Running          Status = "WIP"
)

// compile builds the generated solution code in an isolated, temporary sandbox directory.
// It writes the linked solution file to disk and invokes the configured compiler.
//
// Returns the absolute path to the sandbox directory on success, allowing the caller
// to execute the binary. Returns an error containing the compiler's standard error
// output if compilation fails.
func (m Model) compile() (string, error) {
	log.Debug("Creating sandbox directory for compilation")
	dir, err := os.MkdirTemp("", "sandbox")
	unwrap("unable to create temporary directory", err)

	log.Debug("Sandbox directory created", "path", dir)
	solutionFile := m.getSolution()
	filePath := filepath.Join(dir, m.Compiler.Source)

	log.Debug("Writing solution to sandbox", "file", filePath)
	err = os.WriteFile(filePath, []byte(solutionFile), 0644)
	unwrap("unable to write solution file", err)

	var stderr bytes.Buffer
	log.Info("Starting compilation", "compiler", m.Compile, "args", m.Compiler.Args)
	cmd := exec.Command(m.Compile, m.Compiler.Args...)
	cmd.Dir = dir
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		log.Error("Compilation failed", "stderr", stderr.String(), "err", err)
		return "", errors.New(stderr.String())
	}

	log.Info("Compilation successful")
	return dir, nil
}

// run executes the compiled solution against all provided test cases.
// It enforces strict time limits using context timeouts, pipes test inputs to stdin,
// and captures standard output, standard error, and peak memory usage.
//
// Returns an updated slice of Testcase structs containing the execution results,
// performance metrics, and final statuses (AC, WA, TLE, RE).
func (m Model) run() tea.Msg {
	log.Info("Starting test execution phase")
	tests := m.Tests
	dir, err := m.compile()
	defer func() {
		log.Debug("Cleaning up sandbox directory", "path", dir)
		_ = os.RemoveAll(dir)
	}()

	if err != nil {
		log.Error("Aborting test run due to compilation error")
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
	log.Debug("Parsed run arguments", "args", args)

	wg := &sync.WaitGroup{}
	testsChan := make(chan Testcase, len(tests))

	for i := range tests {
		wg.Add(1)
		go func(test Testcase, ch chan Testcase) {
			defer wg.Done()

			log.Debug("Preparing test case", "index", i)
			// prepare the command for execution
			var stdout, stderr bytes.Buffer
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = dir
			cmd.Stdin = strings.NewReader(test.Input)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			timeout := time.Duration(m.TimeLimit) * time.Millisecond
			log.Debug("Executing test case", "index", i, "timeout", timeout)
			err = cmd.Start()
			unwrap("failed to start program", err)

			var timedOut atomic.Bool
			timer := time.AfterFunc(timeout, func() {
				timedOut.Store(true)
				_ = cmd.Process.Kill()
			})
			defer timer.Stop()

			start := time.Now()
			err = cmd.Wait()
			duration := time.Since(start)

			test.Time = duration.Seconds()
			test.Memory, _ = memory.PeakMemory(cmd)
			test.Output = stdout.String()
			test.Error = stderr.String()

			// determine status based on error and output
			switch {
			case timedOut.Load():
				test.Status = TimeLimitError
				log.Warn("Test case TLE", "index", i, "time", test.Time)
			case err != nil:
				test.Status = RuntimeError
				log.Warn("Test case RE", "index", i, "err", err, "stderr", test.Error)
			case strings.TrimSpace(test.Output) == strings.TrimSpace(test.Answer):
				test.Status = Accepted
				log.Info("Test case AC", "index", i, "time", test.Time, "memory_kb", test.Memory)
			default:
				test.Status = WrongAnswer
				log.Warn("Test case WA", "index", i, "time", test.Time)
			}

			ch <- test
		}(tests[i], testsChan)
	}
	wg.Wait()

	result := make([]Testcase, 0, len(tests))
	for range tests {
		test := <-testsChan
		result = append(result, test)
	}

	finalStatus := getFinalStatus(result)
	log.Info("Finished test execution", "final_status", finalStatus)
	return result
}

// getFinalStatus evaluates a slice of executed test cases and determines the overall
// submission status based on a strict priority hierarchy.
//
// The hierarchy prioritizes critical failures: Compilation Error > Runtime Error >
// Time Limit Exceeded > Wrong Answer > Accepted.
func getFinalStatus(tests []Testcase) Status {
	switch {
	case slices.ContainsFunc(tests,
		func(t Testcase) bool { return t.Status == CompilationError }):
		return CompilationError
	case slices.ContainsFunc(tests,
		func(t Testcase) bool { return t.Status == RuntimeError }):
		return RuntimeError
	case slices.ContainsFunc(tests,
		func(t Testcase) bool { return t.Status == TimeLimitError }):
		return TimeLimitError
	case slices.ContainsFunc(tests,
		func(t Testcase) bool { return t.Status == WrongAnswer }):
		return WrongAnswer
	default:
		return Accepted
	}
}
