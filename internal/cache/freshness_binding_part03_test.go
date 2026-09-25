package cache

import (
	"errors"
	"testing"
)

func TestEvidenceFreshnessC4RequiresDistinctImmutableRefs(t *testing.T) {
	current := evidenceFixture("refs")
	if err := current.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*EvidenceFreshness){
		"missing event ref":    func(e *EvidenceFreshness) { e.EventRef = "" },
		"malformed event ref":  func(e *EvidenceFreshness) { e.EventRef = "push:event" },
		"missing checkout ref": func(e *EvidenceFreshness) { e.CheckoutRef = "" },
		"swapped refs":         func(e *EvidenceFreshness) { e.EventRef, e.CheckoutRef = e.CheckoutRef, e.EventRef },
		"mismatched checkout":  func(e *EvidenceFreshness) { e.CheckoutRef = commitFixtureSHA("other-head") },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := current
			mutate(&mutated)
			if err := mutated.Validate(); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("invalid refs = %v, want ErrInvalidReceipt", err)
			}
			if mutated.Equal(current) {
				t.Fatal("invalid ref mutation retained freshness equality")
			}
		})
	}
}
func TestEvidenceFreshnessRejectsJobsFromAnotherRunOrAttempt(t *testing.T) {
	current := evidenceFixture("run-bound")
	if err := current.Validate(); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*FreshnessJob){
		"run":     func(job *FreshnessJob) { job.RunID = "run-other" },
		"attempt": func(job *FreshnessJob) { job.Attempt++ },
	} {
		t.Run(name, func(t *testing.T) {
			mutated := canonicalEvidence(current)
			job := mutated.Jobs[canonicalTestJob]
			mutate(&job)
			mutated.Jobs[canonicalTestJob] = job
			if err := mutated.Validate(); !errors.Is(err, ErrInvalidReceipt) {
				t.Fatalf("job tuple mutation = %v, want ErrInvalidReceipt", err)
			}
		})
	}
}
