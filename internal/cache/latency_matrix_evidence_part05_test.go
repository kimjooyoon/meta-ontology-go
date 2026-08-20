package cache

import (
	"slices"
	"time"
)

func latencyPercentile(samples []time.Duration, percentile int) time.Duration {
	if len(samples) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), samples...)
	slices.Sort(sorted)
	rank := max((len(sorted)*percentile+99)/100, 1)
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}
