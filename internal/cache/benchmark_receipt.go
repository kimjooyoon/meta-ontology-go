package cache

import "fmt"

// BenchmarkReceipt records reproducible performance evidence separately from
// cache feature or provenance receipts.
type BenchmarkReceipt struct {
	SchemaVersion    string            `json:"schema_version"`
	Fixture          string            `json:"fixture"`
	BaseDigest       Digest            `json:"base_digest"`
	HeadDigest       Digest            `json:"head_digest"`
	RunID            string            `json:"run_id"`
	JobIDs           map[string]string `json:"job_ids"`
	Filesystem       string            `json:"filesystem"`
	Hits             uint64            `json:"hits"`
	Misses           uint64            `json:"misses"`
	Recomputations   uint64            `json:"recomputations"`
	LockWaits        uint64            `json:"lock_waits"`
	P50Nanoseconds   uint64            `json:"p50_nanoseconds"`
	P95Nanoseconds   uint64            `json:"p95_nanoseconds"`
	AllocationsPerOp uint64            `json:"allocations_per_op"`
}

// Validate checks benchmark identity and measurement fields without making a
// performance claim.
func (r BenchmarkReceipt) Validate() error {
	if r.SchemaVersion != benchmarkReceiptSchemaVersion || r.Fixture == "" || !r.BaseDigest.Known() ||
		!r.HeadDigest.Known() || r.RunID == "" || r.Filesystem == "" || len(r.JobIDs) < 6 ||
		r.P50Nanoseconds > r.P95Nanoseconds {
		return fmt.Errorf("%w: incomplete benchmark receipt", ErrInvalidReceipt)
	}
	for name, id := range r.JobIDs {
		if name == "" || id == "" {
			return fmt.Errorf("%w: incomplete benchmark job ID", ErrInvalidReceipt)
		}
	}
	return nil
}
