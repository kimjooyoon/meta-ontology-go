//go:build !linux && !darwin

package bodyexecution

import "os"

func peakRSS(_ *os.ProcessState) *int64 { return nil }
