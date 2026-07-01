//go:build !windows

package memory

import (
	"os/exec"
	"runtime"
	"syscall"
)

func PeakMemory(cmd *exec.Cmd) (uint64, bool) {
	usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0, false
	}

	if runtime.GOOS == "linux" {
		// Linux reports KB
		return uint64(usage.Maxrss), true
	}

	// macOS reports bytes
	return uint64(usage.Maxrss) / 1024, true
}
