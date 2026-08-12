package cache

import "fmt"

// BenchmarkReceipt records reproducible performance evidence separately from
// cache feature or provenance receipts.
type BenchmarkReceipt struct {
	SchemaVersion    string                  `json:"schema_version"`
	Fixture          string                  `json:"fixture"`
	BaseDigest       Digest                  `json:"base_digest"`
	HeadDigest       Digest                  `json:"head_digest"`
	RunID            string                  `json:"run_id"`
	EventID          string                  `json:"event_id"`
	Attempt          uint64                  `json:"attempt"`
	Jobs             map[string]BenchmarkJob `json:"jobs"`
	JobIDs           map[string]string       `json:"job_ids,omitempty"`
	Filesystem       string                  `json:"filesystem"`
	ToolchainDigest  Digest                  `json:"toolchain_digest"`
	PolicyDigest     Digest                  `json:"policy_digest"`
	EvidenceRefs     []EvidenceRef           `json:"evidence_refs"`
	Hits             uint64                  `json:"hits"`
	Misses           uint64                  `json:"misses"`
	Recomputations   uint64                  `json:"recomputations"`
	LockWaits        uint64                  `json:"lock_waits"`
	P50Nanoseconds   uint64                  `json:"p50_nanoseconds"`
	P95Nanoseconds   uint64                  `json:"p95_nanoseconds"`
	AllocationsPerOp uint64                  `json:"allocations_per_op"`
}

// BenchmarkJob binds one canonical CI job to its immutable run result.
type BenchmarkJob struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    Digest `json:"head_sha"`
}

const (
	canonicalCIPolicyJob = "CI policy"
	canonicalSemanticJob = "Semantic conformance"
	canonicalFormatJob   = "gofmt"
	canonicalVetJob      = "go vet"
	canonicalTestJob     = "go test"
	canonicalRaceJob     = "go test -race"
)

var canonicalBenchmarkJobs = []string{
	canonicalCIPolicyJob, canonicalSemanticJob, canonicalFormatJob,
	canonicalVetJob, canonicalTestJob, canonicalRaceJob,
}

// Validate checks benchmark identity and measurement fields without making a
// performance claim.
func (r BenchmarkReceipt) Validate() error {
	if r.SchemaVersion != benchmarkReceiptSchemaVersion || r.Fixture == "" || !r.BaseDigest.Known() ||
		!r.HeadDigest.Known() || r.RunID == "" || r.EventID == "" || r.Attempt == 0 ||
		r.Filesystem == "" || !r.ToolchainDigest.Known() || !r.PolicyDigest.Known() ||
		r.JobIDs != nil || len(r.Jobs) != len(canonicalBenchmarkJobs) ||
		r.P50Nanoseconds > r.P95Nanoseconds {
		return fmt.Errorf("%w: incomplete benchmark receipt", ErrInvalidReceipt)
	}
	if err := validateEvidenceRefs(r.EvidenceRefs); err != nil {
		return err
	}
	if err := validateEvidenceBindings(EvidenceFreshness{
		PolicyDigest: r.PolicyDigest, ToolchainDigest: r.ToolchainDigest,
		EvidenceRefs: r.EvidenceRefs,
	}); err != nil {
		return err
	}
	seenIDs := make(map[string]struct{}, len(r.Jobs))
	for _, name := range canonicalBenchmarkJobs {
		job, exists := r.Jobs[name]
		if !exists || job.ID == "" || job.Status == "" || job.Conclusion == "" ||
			!job.HeadSHA.Known() || job.HeadSHA != r.HeadDigest {
			return fmt.Errorf("%w: incomplete benchmark job %q", ErrInvalidReceipt, name)
		}
		if _, exists := seenIDs[job.ID]; exists {
			return fmt.Errorf("%w: replayed benchmark job ID %q", ErrInvalidReceipt, job.ID)
		}
		seenIDs[job.ID] = struct{}{}
	}
	return nil
}
