package resourceenvelope

import (
	"fmt"
)

// Validate checks the non-observational envelope shape. It does not evaluate
// resource limits or require a non-zero denominator.
func (e Envelope) Validate() error {
	if e.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if e.RunnerImageDigest == "" {
		return fmt.Errorf("runner_image_digest is required")
	}
	if e.AllocatedCPUCount == 0 {
		return fmt.Errorf("allocated_cpu_count must be positive")
	}
	if e.WarmupCount != ExpectedWarmupCount {
		return fmt.Errorf("warmup_count must be %d", ExpectedWarmupCount)
	}
	if e.SampleCount != ExpectedSampleCount {
		return fmt.Errorf("sample_count must be %d", ExpectedSampleCount)
	}
	if len(e.Samples) != int(ExpectedWarmupCount+ExpectedSampleCount) {
		return fmt.Errorf("samples must contain one warmup and %d measured observations", ExpectedSampleCount)
	}
	return nil
}

// Validate checks that a result is sealed and internally coherent.
func (r Result) Validate() error {
	if r.SchemaVersion != SchemaVersion {
		return fmt.Errorf("schema_version must be %q", SchemaVersion)
	}
	if r.Status != PASS && r.Status != FAIL_CLOSED && r.Status != UNKNOWN {
		return fmt.Errorf("unsupported status %q", r.Status)
	}
	if r.CanonicalDigest == "" || r.CanonicalDigest != r.computedDigest() {
		return fmt.Errorf("canonical_digest is invalid")
	}
	if r.Digest != "" && r.Digest != r.CanonicalDigest {
		return fmt.Errorf("digest alias does not match canonical_digest")
	}
	return nil
}
