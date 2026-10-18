//go:build darwin

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
	if !ok {
		return 0, fmt.Errorf("peak RSS is unavailable")
	}
	return peakRSSUsage(usage)
}

// CurrentPeakRSSKiB observes this process through the operating system.
func CurrentPeakRSSKiB() (int64, error) {
	var usage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &usage); err != nil {
		return 0, err
	}
	return peakRSSUsage(&usage)
}

func peakRSSUsage(usage *syscall.Rusage) (int64, error) {
	if usage == nil || usage.Maxrss <= 0 {
		return 0, fmt.Errorf("peak RSS is unavailable")
	}
	return (usage.Maxrss + 1023) / 1024, nil
}
