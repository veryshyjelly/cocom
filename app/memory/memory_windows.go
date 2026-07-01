//go:build windows

package memory

import "os/exec"

func PeakMemory(cmd *exec.Cmd) (uint64, bool) {
	return 0, false
}
