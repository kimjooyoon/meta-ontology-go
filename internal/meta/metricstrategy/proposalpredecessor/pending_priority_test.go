package proposalpredecessor

import (
	"strings"
	"testing"
)

func TestPendingObservationPreservesSelectionAndRefutationPriority(t *testing.T) {
	head, current := strings.Repeat("a", 40), strings.Repeat("b", 40)
	pending := githubRun{ID: 42, RunAttempt: 1, HeadSHA: head,
		HeadBranch: RouteDev, Event: "push", Status: "queued", Name: workflowName}
	base := Collection{RequestedRoute: RouteDev, ObservedRuns: 1, ExactRuns: 1,
		Unresolved: 1, pending: []githubRun{pending}}
	tests := []struct {
		name   string
		edit   func(*Collection)
		reason string
		await  bool
	}{
		{"pending", func(*Collection) {}, ReasonEvidenceUnknown, true},
		{"refuted", func(c *Collection) { c.Contradictions = 1 }, ReasonRouteContradiction, false},
		{"ambiguous", func(c *Collection) {
			c.Candidates = []Candidate{readyCandidate(head, 1), readyCandidate(head, 2)}
		}, ReasonAmbiguous, false},
		{"mixed-unknown", func(c *Collection) { c.Unresolved++ }, ReasonNotFound, false},
		{"wrong-route", func(c *Collection) {
			c.pending = []githubRun{pending}
			c.pending[0].HeadBranch = RouteMain
		}, ReasonNotFound, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			collection := base
			test.edit(&collection)
			report, payload, err := SelectPending("owner/repo", current, head, collection)
			if err == nil || len(payload) != 0 || report.Reason != test.reason ||
				AwaitablePending(report) != test.await {
				t.Fatalf("report=%+v payload=%q err=%v", report, payload, err)
			}
		})
	}
	candidate := readyCandidate(head, 7)
	closed := Collection{RequestedRoute: RouteDev, ObservedRuns: 1, ExactRuns: 1,
		ObservedJobs: 1, ExactJobs: 1, ObservedArtifacts: 1, ExactArtifacts: 1,
		Candidates: []Candidate{candidate}}
	report, _, err := SelectPending("owner/repo", current, head, closed)
	if err != nil || !report.Ready() || report.Summary.ProofsPassed != 5 || AwaitablePending(report) {
		t.Fatalf("completed predecessor did not close normally: %+v %v", report, err)
	}
}
