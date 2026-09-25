package languageprofile

import (
	"runtime"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/sourceexecution"
)

type RuntimeMeasurer struct{}

func (RuntimeMeasurer) Measure(run func() sourceexecution.Receipt) (sourceexecution.Receipt, Measurement) {
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	started := time.Now()
	receipt := run()
	wall := time.Since(started).Nanoseconds()
	runtime.ReadMemStats(&after)
	allocated := uint64(0)
	if after.TotalAlloc >= before.TotalAlloc {
		allocated = after.TotalAlloc - before.TotalAlloc
	}
	return receipt, Measurement{WallNanoseconds: wall, TotalAllocBytes: allocated}
}
