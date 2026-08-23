package proposalpredecessor

import (
	"strings"
	"testing"
)

func TestFailedWorkflowWithSuccessfulSynthesisIsSelectable(t *testing.T) {
	predecessor := strings.Repeat("e", 40)
	candidate := readyCandidate(predecessor, 9)
	if candidate.Conclusion != "failure" {
		t.Fatal("fixture must prove job-level resolution")
	}
	report, _, err := Select(
		"owner/repository", strings.Repeat("f", 40), predecessor,
		Collection{ObservedRuns: 1, ExactRuns: 1, ObservedJobs: 5,
			ExactJobs: 1, ObservedArtifacts: 4, ExactArtifacts: 1,
			Candidates: []Candidate{candidate}},
	)
	if err != nil || !report.Ready() || report.Summary.ExactJobs != 1 {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestFailedSynthesisJobFailsClosed(t *testing.T) {
	predecessor := strings.Repeat("1", 40)
	candidate := readyCandidate(predecessor, 10)
	candidate.SynthesisJobConclusion = "failure"
	report, _, err := Select(
		"owner/repository", strings.Repeat("2", 40), predecessor,
		Collection{ObservedRuns: 1, ExactRuns: 1, ObservedJobs: 5,
			ExactJobs: 1, ObservedArtifacts: 4, ExactArtifacts: 1,
			Candidates: []Candidate{candidate}},
	)
	if err == nil || report.Decision != "FAIL_CLOSED" {
		t.Fatalf("report=%+v err=%v", report, err)
	}
}

func TestSynthesisJobCardinalityFailsClosed(t *testing.T) {
	for _, exactJobs := range []int{0, 2} {
		predecessor := strings.Repeat("3", 40)
		report, _, err := Select(
			"owner/repository", strings.Repeat("4", 40), predecessor,
			Collection{ObservedRuns: 1, ExactRuns: 1, ObservedJobs: exactJobs,
				ExactJobs: exactJobs, ObservedArtifacts: 1, ExactArtifacts: 1,
				Candidates: []Candidate{readyCandidate(predecessor, 11)}},
		)
		if err == nil || report.Decision != "FAIL_CLOSED" ||
			report.Reason != "PROPOSAL_SYNTHESIS_JOB_CARDINALITY" {
			t.Fatalf("exact_jobs=%d report=%+v err=%v", exactJobs, report, err)
		}
	}
}
