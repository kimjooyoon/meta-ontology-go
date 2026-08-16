package cache

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestIncrementalMeasurementReceiptIsMachineReadableAndDeterministic(t *testing.T) {
	receipt, binding := validIncrementalMeasurementReceipt()
	if err := receipt.Validate(binding); err != nil {
		t.Fatal(err)
	}
	firstDigest, err := receipt.StableDigest()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	var decoded IncrementalMeasurementReceipt
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, receipt) {
		t.Fatalf("JSON round-trip changed receipt: %#v != %#v", decoded, receipt)
	}
	secondDigest, err := decoded.StableDigest()
	if err != nil || secondDigest != firstDigest {
		t.Fatalf("JSON round-trip digest = %s, %v; want %s", secondDigest, err, firstDigest)
	}
	if err := decoded.Validate(binding); err != nil {
		t.Fatal(err)
	}
}

func TestIncrementalMeasurementReceiptFailsClosedOnIdentity(t *testing.T) {
	receipt, binding := validIncrementalMeasurementReceipt()
	for name, mutate := range map[string]func(*IncrementalMeasurementReceipt, *IncrementalMeasurementBinding){
		"missing receipt baseline": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.BaselineDigest = "" },
		"mismatched receipt baseline": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) {
			r.BaselineDigest = HashBytes([]byte("other-baseline"))
		},
		"missing expected baseline": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) { b.BaselineDigest = "" },
		"mismatched expected baseline": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) {
			b.BaselineDigest = HashBytes([]byte("other-baseline"))
		},
		"missing options": func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.OptionsDigest = "" },
		"mismatched options": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) {
			b.OptionsDigest = HashBytes([]byte("other-options"))
		},
		"missing Go version":   func(r *IncrementalMeasurementReceipt, _ *IncrementalMeasurementBinding) { r.GoVersion = "" },
		"mismatched toolchain": func(_ *IncrementalMeasurementReceipt, b *IncrementalMeasurementBinding) { b.Toolchain = "go-other" },
	} {
		t.Run(name, func(t *testing.T) {
			mutatedReceipt, mutatedBinding := cloneIncrementalReceipt(receipt, binding)
			mutate(&mutatedReceipt, &mutatedBinding)
			if err := mutatedReceipt.Validate(mutatedBinding); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("identity mutation = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}

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
