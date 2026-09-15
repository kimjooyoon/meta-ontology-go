package performance

import "testing"

// BenchmarkStage adapts a StageSpec to Go's benchmark runner. The standard
// benchmark output includes wall-clock data for diagnosis; the custom ops/op
// metric is the stable signal used by budget checks.
func BenchmarkStage(b *testing.B, spec StageSpec) {
	if err := spec.validate(); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	var total uint64
	var counter Counter
	for i := 0; i < b.N; i++ {
		counter.reset()
		if err := spec.Run(&counter); err != nil {
			b.Fatal(err)
		}
		total += counter.Operations()
	}
	if b.N > 0 {
		b.ReportMetric(float64(total)/float64(b.N), "ops/op")
	}
}
