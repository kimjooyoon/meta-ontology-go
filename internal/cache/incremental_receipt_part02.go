package cache

import (
	"fmt"
)

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
