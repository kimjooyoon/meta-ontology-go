package cache

import (
	"fmt"
)

// Validate checks benchmark identity and measurement fields without making a
// performance claim.
func (r BenchmarkReceipt) Validate() error {
	if r.SchemaVersion != benchmarkReceiptSchemaVersion || r.Fixture == "" || !r.BaseDigest.Known() ||
		!r.HeadDigest.Known() || !validCommitSHA(r.BaseSHA) || !validCommitSHA(r.HeadSHA) ||
		r.Event != "pull_request" || r.Workflow == "" || r.WorkflowRunID == "" ||
		r.RunID == "" || r.EventID == "" || r.Attempt == 0 ||
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
		if !exists || job.ID == "" || job.RunID != r.WorkflowRunID || job.Attempt != r.Attempt ||
			job.Status != "completed" || job.Conclusion != "success" ||
			!job.HeadSHA.Known() || job.HeadSHA != r.HeadDigest || job.HeadCommitSHA != r.HeadSHA {
			return fmt.Errorf("%w: incomplete benchmark job %q", ErrInvalidReceipt, name)
		}
		if _, exists := seenIDs[job.ID]; exists {
			return fmt.Errorf("%w: replayed benchmark job ID %q", ErrInvalidReceipt, job.ID)
		}
		seenIDs[job.ID] = struct{}{}
	}
	return nil
}
