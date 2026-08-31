//go:build !linux && !darwin

package toolchaincli

import (
	"fmt"
	"os"
)

func peakRSSKiB(state *os.ProcessState) (int64, error) {
	return 0, fmt.Errorf("peak RSS observation is unsupported on this runner")
}
