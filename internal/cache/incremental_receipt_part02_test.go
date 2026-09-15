package cache

import (
	"errors"
	"testing"
)

func TestIncrementalMeasurementReceiptRejectsMatrixMutations(t *testing.T) {
	receipt, binding := validIncrementalMeasurementReceipt()
	mutations := map[string]func(*IncrementalMeasurementReceipt){
		"wrong schema":         func(r *IncrementalMeasurementReceipt) { r.SchemaVersion = "v0" },
		"missing fixture size": func(r *IncrementalMeasurementReceipt) { r.FixtureSizes = nil },
		"wrong fixture size":   func(r *IncrementalMeasurementReceipt) { r.FixtureSizes[0] = 11 },
		"reordered fixture sizes": func(r *IncrementalMeasurementReceipt) {
			r.FixtureSizes[0], r.FixtureSizes[1] = r.FixtureSizes[1], r.FixtureSizes[0]
		},
		"missing measurement": func(r *IncrementalMeasurementReceipt) { r.Measurements = r.Measurements[:len(r.Measurements)-1] },
		"reordered measurements": func(r *IncrementalMeasurementReceipt) {
			r.Measurements[0], r.Measurements[1] = r.Measurements[1], r.Measurements[0]
		},
		"wrong mutation class": func(r *IncrementalMeasurementReceipt) { r.Measurements[0].MutationClass = "unknown" },
		"hit miss count":       func(r *IncrementalMeasurementReceipt) { r.Measurements[0].Hits = 9 },
		"recomputation count":  func(r *IncrementalMeasurementReceipt) { r.Measurements[1].Recomputations = 0 },
		"missing sample":       func(r *IncrementalMeasurementReceipt) { r.Measurements[0].SampleCount = 0 },
		"p95 before p50":       func(r *IncrementalMeasurementReceipt) { r.Measurements[0].P95Nanoseconds = 0 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			mutated, _ := cloneIncrementalReceipt(receipt, binding)
			mutate(&mutated)
			if err := mutated.Validate(binding); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("matrix mutation = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}
func validIncrementalMeasurementReceipt() (IncrementalMeasurementReceipt, IncrementalMeasurementBinding) {
	binding := IncrementalMeasurementBinding{
		BaselineDigest: HashBytes([]byte("incremental-baseline")), GoVersion: "go1.26.5",
		Toolchain: "go1.26.5", OptionsDigest: HashBytes([]byte("incremental-options")),
	}
	measurements := make([]IncrementalMeasurement, 0, len(canonicalIncrementalFixtureSizes)*len(canonicalIncrementalMutations))
	for _, size := range canonicalIncrementalFixtureSizes {
		for _, mutation := range canonicalIncrementalMutations {
			hits, misses := uint64(incrementalMeasurementPartCount), uint64(0)
			if mutation != "presentation_rename" {
				hits, misses = incrementalMeasurementPartCount-1, 1
			}
			measurements = append(measurements, IncrementalMeasurement{
				FixtureSize: size, MutationClass: mutation, Hits: hits, Misses: misses,
				Recomputations: misses, P50Nanoseconds: size, P95Nanoseconds: size * 2, SampleCount: 5,
			})
		}
	}
	return IncrementalMeasurementReceipt{
		SchemaVersion: incrementalMeasurementReceiptSchemaVersion, FixtureSizes: append([]uint64(nil), canonicalIncrementalFixtureSizes...),
		Measurements: measurements, GoVersion: binding.GoVersion, Toolchain: binding.Toolchain,
		OptionsDigest: binding.OptionsDigest, BaselineDigest: binding.BaselineDigest,
	}, binding
}
func cloneIncrementalReceipt(receipt IncrementalMeasurementReceipt, binding IncrementalMeasurementBinding) (IncrementalMeasurementReceipt, IncrementalMeasurementBinding) {
	receipt.FixtureSizes = append([]uint64(nil), receipt.FixtureSizes...)
	receipt.Measurements = append([]IncrementalMeasurement(nil), receipt.Measurements...)
	return receipt, binding
}
