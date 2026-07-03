//go:build windows

package memory

import "os/exec"

// PeakMemory retrieves the peak resident set size (RSS) memory usage of an
// executed command from the OS process state.
//
// Returns the memory usage in kilobytes and a boolean indicating whether the
// metric was successfully retrieved. (Note: The Windows implementation currently
// returns 0 and false as a placeholder).
func PeakMemory(cmd *exec.Cmd) (uint64, bool) {
	return 0, false
}
