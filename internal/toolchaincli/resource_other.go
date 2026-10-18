//go:build !linux && !darwin

package toolchaincli

import (
	"fmt"
	"os"
)

func peakRSSKiB(state *os.ProcessState) (int64, error) {
	return 0, fmt.Errorf("peak RSS observation is unsupported on this runner")
}

// CurrentPeakRSSKiB does not invent a measurement on unsupported platforms.
func CurrentPeakRSSKiB() (int64, error) {
	return 0, fmt.Errorf("peak RSS observation is unsupported on this runner")
}
