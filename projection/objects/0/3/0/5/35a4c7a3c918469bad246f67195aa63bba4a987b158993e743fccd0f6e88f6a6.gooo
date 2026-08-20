package cache

// BenchmarkReceipt records reproducible performance evidence separately from
// cache feature or provenance receipts.
type BenchmarkReceipt struct {
	SchemaVersion    string                  `json:"schema_version"`
	Fixture          string                  `json:"fixture"`
	BaseDigest       Digest                  `json:"base_digest"`
	HeadDigest       Digest                  `json:"head_digest"`
	BaseSHA          string                  `json:"base_sha"`
	HeadSHA          string                  `json:"head_sha"`
	Event            string                  `json:"event"`
	Workflow         string                  `json:"workflow"`
	WorkflowRunID    string                  `json:"workflow_run_id"`
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
	ID            string `json:"id"`
	RunID         string `json:"run_id"`
	Attempt       uint64 `json:"attempt"`
	Status        string `json:"status"`
	Conclusion    string `json:"conclusion"`
	HeadSHA       Digest `json:"head_sha"`
	HeadCommitSHA string `json:"head_commit_sha"`
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
