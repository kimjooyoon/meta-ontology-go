//go:build linux

package toolchaincli

import (
	"fmt"
	"os"
	"syscall"
)

func peakRSSKiB(state *os.ProcessState) (int64, error) {
	if state == nil {
		return 0, fmt.Errorf("process resource state is unavailable")
	}
	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok || usage == nil || usage.Maxrss <= 0 {
		return 0, fmt.Errorf("peak RSS is unavailable")
	}
	return usage.Maxrss, nil
}
