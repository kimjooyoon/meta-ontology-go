package resourceenvelope

import "fmt"

// SchemaVersion identifies the resource-envelope wire contract.
const SchemaVersion = "gooo/resource-envelope/v1"

const (
	ExpectedWarmupCount uint64 = 1
	ExpectedSampleCount uint64 = 5
)

// Status is the conservative outcome of an evaluation.
type Status string

const (
	PASS        Status = "PASS"
	FAIL_CLOSED Status = "FAIL_CLOSED"
	UNKNOWN     Status = "UNKNOWN"

	StatusPass       = PASS
	StatusFailClosed = FAIL_CLOSED
	StatusUnknown    = UNKNOWN
)

// EvaluationStatus is an alias retained for callers that prefer the longer
// name at API boundaries.
type EvaluationStatus = Status

// Envelope is the strict input contract. Samples contains one warmup sample
// followed by SampleCount measured samples.
type Envelope struct {
	SchemaVersion     string   `json:"schema_version"`
	RunnerImageDigest string   `json:"runner_image_digest"`
	AllocatedCPUCount uint64   `json:"allocated_cpu_count"`
	WarmupCount       uint64   `json:"warmup_count"`
	SampleCount       uint64   `json:"sample_count"`
	Limits            Limits   `json:"limits"`
	Samples           []Sample `json:"samples"`
}

// Limits are inclusive upper bounds for the derived resource values.
type Limits struct {
	CPUCoreNS    uint64 `json:"cpu_core_ns"`
	PeakRSSBytes uint64 `json:"peak_rss_bytes"`
	ReadBytes    uint64 `json:"read_bytes"`
	WriteBytes   uint64 `json:"write_bytes"`
}

// Sample is one unsigned observation supplied by a runner.
type Sample struct {
	CPUCoreNS    uint64 `json:"cpu_core_ns"`
	WallNS       uint64 `json:"wall_ns"`
	PeakRSSBytes uint64 `json:"peak_rss_bytes"`
	ReadBytes    uint64 `json:"read_bytes"`
	WriteBytes   uint64 `json:"write_bytes"`
}

// Result is the canonical integer-only evaluation result. CanonicalDigest is
// SHA-256 over the same result with canonical_digest set to the empty string.
type Result struct {
	CPUCoreNS         uint64 `json:"cpu_core_ns"`
	CPUUtilizationPPM uint64 `json:"cpu_utilization_ppm"`
	CanonicalDigest   string `json:"canonical_digest"`
	PeakRSSBytes      uint64 `json:"peak_rss_bytes"`
	ReadBytes         uint64 `json:"read_bytes"`
	SchemaVersion     string `json:"schema_version"`
	Status            Status `json:"status"`
	WriteBytes        uint64 `json:"write_bytes"`

	// Digest is a Go-side compatibility alias. It is not part of the wire
	// result, so it cannot change canonical bytes.
	Digest     string `json:"-"`
	ReasonCode string `json:"-"`
}

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
