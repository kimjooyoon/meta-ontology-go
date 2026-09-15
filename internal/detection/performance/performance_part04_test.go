package performance

import (
	"testing"
)

func fixedWork(operations uint64) Runner {
	return func(counter *Counter) error {
		counter.Add(operations)
		return nil
	}
}
func BenchmarkSyntheticPipeline(b *testing.B) {
	for _, spec := range syntheticSpecs() {
		b.Run(string(spec.Stage), func(b *testing.B) {
			BenchmarkStage(b, spec)
		})
	}
}
