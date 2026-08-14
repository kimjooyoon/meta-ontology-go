package cache

import "fmt"

const (
	incrementalMeasurementReceiptSchemaVersion = "v1"
	incrementalMeasurementPartCount            = 10
)

var (
	canonicalIncrementalFixtureSizes = []uint64{10, 100, 1000, 10000}
	canonicalIncrementalMutations    = []string{"presentation_rename", "local_fact", "dependency_closure"}
)

// IncrementalMeasurementBinding identifies the environment against which a
// measurement receipt is allowed to be consumed. Baseline and options are
// content digests so a missing or replayed environment fails closed.
type IncrementalMeasurementBinding struct {
	BaselineDigest Digest
	GoVersion      string
	Toolchain      string
	OptionsDigest  Digest
}

// IncrementalMeasurement records one deterministic fixture/mutation result.
// Timing values are evidence only; validation does not impose a performance
// threshold.
type IncrementalMeasurement struct {
	FixtureSize    uint64 `json:"fixture_size"`
	MutationClass  string `json:"mutation_class"`
	Hits           uint64 `json:"hits"`
	Misses         uint64 `json:"misses"`
	Recomputations uint64 `json:"recomputations"`
	P50Nanoseconds uint64 `json:"p50_nanoseconds"`
	P95Nanoseconds uint64 `json:"p95_nanoseconds"`
	SampleCount    uint64 `json:"sample_count"`
}

// IncrementalMeasurementReceipt is a machine-readable receipt for the
// incremental cache evidence matrix. It is deliberately separate from
// BenchmarkReceipt, whose identity is one external CI run and one fixture.
type IncrementalMeasurementReceipt struct {
	SchemaVersion  string                   `json:"schema_version"`
	FixtureSizes   []uint64                 `json:"fixture_sizes"`
	Measurements   []IncrementalMeasurement `json:"measurements"`
	GoVersion      string                   `json:"go_version"`
	Toolchain      string                   `json:"toolchain"`
	OptionsDigest  Digest                   `json:"options_digest"`
	BaselineDigest Digest                   `json:"baseline_digest"`
}

// Validate binds the receipt to the caller's exact baseline and environment.
// A caller must provide a complete binding; an absent or mismatched baseline
// is never treated as an advisory omission.
func (r IncrementalMeasurementReceipt) Validate(binding IncrementalMeasurementBinding) error {
	if err := binding.validate(); err != nil {
		return err
	}
	if r.SchemaVersion != incrementalMeasurementReceiptSchemaVersion ||
		r.GoVersion != binding.GoVersion || r.Toolchain != binding.Toolchain ||
		r.OptionsDigest != binding.OptionsDigest || r.BaselineDigest != binding.BaselineDigest {
		return fmt.Errorf("%w: incremental measurement identity mismatch", ErrInvalidReceipt)
	}
	if !equalUint64s(r.FixtureSizes, canonicalIncrementalFixtureSizes) ||
		len(r.Measurements) != len(canonicalIncrementalFixtureSizes)*len(canonicalIncrementalMutations) {
		return fmt.Errorf("%w: incomplete incremental fixture matrix", ErrInvalidReceipt)
	}
	measurementIndex := 0
	for _, size := range canonicalIncrementalFixtureSizes {
		for _, mutation := range canonicalIncrementalMutations {
			measurement := r.Measurements[measurementIndex]
			if measurement.FixtureSize != size || measurement.MutationClass != mutation {
				return fmt.Errorf("%w: non-canonical incremental measurement order", ErrInvalidReceipt)
			}
			if measurement.Hits > incrementalMeasurementPartCount || measurement.Misses > incrementalMeasurementPartCount ||
				measurement.Hits+measurement.Misses != incrementalMeasurementPartCount ||
				measurement.Recomputations != measurement.Misses || measurement.SampleCount == 0 ||
				measurement.P50Nanoseconds > measurement.P95Nanoseconds {
				return fmt.Errorf("%w: invalid incremental measurement %d", ErrInvalidReceipt, measurementIndex)
			}
			measurementIndex++
		}
	}
	return nil
}

// StableDigest returns a deterministic content digest for a valid JSON
// receipt value. The canonical encoder preserves measurement order and map
// independence for callers that persist or compare the receipt.
func (r IncrementalMeasurementReceipt) StableDigest() (Digest, error) {
	return DigestOf(r)
}

func (b IncrementalMeasurementBinding) validate() error {
	if !b.BaselineDigest.Known() || !b.OptionsDigest.Known() || b.GoVersion == "" || b.Toolchain == "" {
		return fmt.Errorf("%w: incomplete incremental measurement binding", ErrInvalidReceipt)
	}
	return nil
}

func equalUint64s(left, right []uint64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
