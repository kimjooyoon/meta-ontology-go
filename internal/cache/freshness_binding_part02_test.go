package cache

import (
	"testing"
)

func TestEvidenceFreshnessC4RejectsStaleAndReplayTuples(t *testing.T) {
	current := evidenceFixture("run-current")
	for name, mutate := range map[string]func(*EvidenceFreshness){
		"base":         func(e *EvidenceFreshness) { e.BaseDigest = HashBytes([]byte("new-base")) },
		"head":         func(e *EvidenceFreshness) { e.HeadDigest = HashBytes([]byte("new-head")) },
		"base sha":     func(e *EvidenceFreshness) { e.BaseSHA = commitFixtureSHA("new-base") },
		"head sha":     func(e *EvidenceFreshness) { e.HeadSHA = commitFixtureSHA("new-head") },
		"run":          func(e *EvidenceFreshness) { e.RunID = "run-other" },
		"event kind":   func(e *EvidenceFreshness) { e.Event = "push" },
		"event ref":    func(e *EvidenceFreshness) { e.EventRef = "refs/pull/9/merge" },
		"checkout ref": func(e *EvidenceFreshness) { e.CheckoutRef = commitFixtureSHA("other-head") },
		"event":        func(e *EvidenceFreshness) { e.EventID = "event-other" },
		"attempt":      func(e *EvidenceFreshness) { e.Attempt++ },
		"job run": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.RunID = "run-other"
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job attempt": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.Attempt++
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job head": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.HeadSHA = commitFixtureSHA("other-head")
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job status": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.Status = "in_progress"
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"job conclusion": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.Conclusion = "failure"
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"duplicate job id": func(e *EvidenceFreshness) {
			jobs := copyFreshnessJobs(e.Jobs)
			job := jobs[canonicalTestJob]
			job.ID = jobs[canonicalCIPolicyJob].ID
			jobs[canonicalTestJob] = job
			e.Jobs = jobs
		},
		"prior": func(e *EvidenceFreshness) { e.PredecessorDigests[0] = HashBytes([]byte("other")) },
	} {
		t.Run(name, func(t *testing.T) {
			stale := canonicalEvidence(current)
			mutate(&stale)
			if stale.Matches(current) {
				t.Fatal("stale evidence matched current tuple")
			}
		})
	}
}
