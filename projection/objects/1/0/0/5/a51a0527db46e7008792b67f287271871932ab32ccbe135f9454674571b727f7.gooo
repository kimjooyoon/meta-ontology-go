package cache

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
)

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
	maps.Copy(copyOf, jobs)
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
	slices.Sort(e.PredecessorDigests)
	e.EvidenceRefs = append([]EvidenceRef(nil), e.EvidenceRefs...)
	sort.Slice(e.EvidenceRefs, func(i, j int) bool { return e.EvidenceRefs[i].Name < e.EvidenceRefs[j].Name })
	return e
}
