package cache

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
