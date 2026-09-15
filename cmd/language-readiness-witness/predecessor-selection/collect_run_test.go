package main

import "testing"

func TestFindProducerJobPreservesExactCardinality(t *testing.T) {
	jobs := []workflowJob{
		{ID: 1, Name: "guarded-promotion-receipt", Conclusion: "failure"},
		{ID: 2, Name: "language-concept-artifact", Status: "completed",
			Conclusion: "success", RunAttempt: 1},
	}
	job, matches := findProducerJob(jobs, "language-concept-artifact")
	if matches != 1 || job.ID != 2 {
		t.Fatalf("producer = %+v, matches = %d", job, matches)
	}
	jobs = append(jobs, workflowJob{ID: 3, Name: "language-concept-artifact"})
	if _, matches = findProducerJob(jobs, "language-concept-artifact"); matches != 2 {
		t.Fatalf("duplicate producer matches = %d, want 2", matches)
	}
}
