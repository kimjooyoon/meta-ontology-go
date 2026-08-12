package cache

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

// FreshnessJob binds one required PR check to the exact workflow head. It is
// separate from content digests because Git commit IDs are identities, not
// SHA-256 digests of durable payloads.
type FreshnessJob struct {
	ID         string `json:"id"`
	RunID      string `json:"run_id"`
	Attempt    uint64 `json:"attempt"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSHA    string `json:"head_sha"`
}

func validCommitSHA(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	if value != strings.ToLower(value) {
		return false
	}
	if _, err := hex.DecodeString(value); err != nil {
		return false
	}
	return value != strings.Repeat("0", len(value))
}

func validEventRef(value string) bool {
	parts := strings.Split(value, "/")
	if len(parts) != 4 || parts[0] != "refs" || parts[1] != "pull" ||
		(parts[3] != "head" && parts[3] != "merge") || strings.TrimSpace(value) != value {
		return false
	}
	if parts[2] == "" || parts[2] == "0" {
		return false
	}
	for _, char := range parts[2] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func validateFreshnessRefs(eventRef, checkoutRef, headSHA string) error {
	if !validEventRef(eventRef) {
		return fmt.Errorf("%w: malformed event ref", ErrInvalidReceipt)
	}
	if !validCommitSHA(checkoutRef) {
		return fmt.Errorf("%w: malformed checkout ref", ErrInvalidReceipt)
	}
	if checkoutRef != headSHA {
		return fmt.Errorf("%w: checkout ref does not match head SHA", ErrInvalidReceipt)
	}
	return nil
}

func validateFreshnessJobs(jobs map[string]FreshnessJob, runID string, attempt uint64, headSHA string) error {
	if len(jobs) != len(canonicalBenchmarkJobs) {
		return fmt.Errorf("%w: incomplete canonical CI jobs", ErrInvalidReceipt)
	}
	seenIDs := make(map[string]struct{}, len(jobs))
	for _, name := range canonicalBenchmarkJobs {
		job, ok := jobs[name]
		if !ok || strings.TrimSpace(job.ID) == "" || job.RunID != runID || job.Attempt != attempt ||
			job.Status != "completed" ||
			job.Conclusion != "success" || job.HeadSHA != headSHA {
			return fmt.Errorf("%w: non-terminal or mismatched canonical CI job %q", ErrInvalidReceipt, name)
		}
		if _, exists := seenIDs[job.ID]; exists {
			return fmt.Errorf("%w: replayed canonical CI job ID %q", ErrInvalidReceipt, job.ID)
		}
		seenIDs[job.ID] = struct{}{}
	}
	return nil
}

func copyFreshnessJobs(jobs map[string]FreshnessJob) map[string]FreshnessJob {
	if jobs == nil {
		return nil
	}
	copyOf := make(map[string]FreshnessJob, len(jobs))
	for name, job := range jobs {
		copyOf[name] = job
	}
	return copyOf
}

func freshnessJobsEqual(left, right map[string]FreshnessJob) bool {
	if len(left) != len(right) {
		return false
	}
	for name, job := range left {
		if right[name] != job {
			return false
		}
	}
	return true
}

func canonicalEvidence(e EvidenceFreshness) EvidenceFreshness {
	e.Jobs = copyFreshnessJobs(e.Jobs)
	e.PredecessorDigests = append([]Digest(nil), e.PredecessorDigests...)
	sort.Slice(e.PredecessorDigests, func(i, j int) bool { return e.PredecessorDigests[i] < e.PredecessorDigests[j] })
	e.EvidenceRefs = append([]EvidenceRef(nil), e.EvidenceRefs...)
	sort.Slice(e.EvidenceRefs, func(i, j int) bool { return e.EvidenceRefs[i].Name < e.EvidenceRefs[j].Name })
	return e
}
