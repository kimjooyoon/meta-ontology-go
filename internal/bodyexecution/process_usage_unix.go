//go:build linux || darwin

package bodyexecution

import (
	"os"
	"runtime"
	"syscall"
)

func peakRSS(state *os.ProcessState) *int64 {
	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok {
		return nil
	}
	value := usage.Maxrss
	if runtime.GOOS == "linux" {
		value *= 1024
	}
	return &value
}
