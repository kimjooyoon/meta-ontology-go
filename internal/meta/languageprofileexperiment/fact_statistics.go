package languageprofileexperiment

import (
	"slices"
	"strings"
)

func int64Stats(values []int64) (int64, int64, int64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	slices.Sort(values)
	return values[0], values[len(values)/2], values[len(values)-1]
}

func uint64Stats(values []uint64) (uint64, uint64, uint64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	slices.Sort(values)
	return values[0], values[len(values)/2], values[len(values)-1]
}

func validDigest(value string) bool { return len(value) == 71 && strings.HasPrefix(value, "sha256:") }

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
