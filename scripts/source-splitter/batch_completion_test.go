package main

import "testing"

func TestBatchKeepsSuccessorProjectionPendingInItsOutcome(t *testing.T) {
	for _, selected := range []int{0, 1} {
		plans := make([]plannedSplit, selected)
		subjects := make([]splitBatchSubject, selected)
		report := buildSplitBatchReport("source", "plan", plans, subjects, 1, nil)
		if report.Decision != "PROGRESS" || report.Resolution != "LOGICAL_ONLY" ||
			report.Reason != "SUCCESSOR_PROJECTION_REQUIRED" {
			t.Fatalf("pending physical work was hidden by the logical outcome: %s/%s/%s",
				report.Decision, report.Resolution, report.Reason)
		}
		if !report.Exact || !report.SuccessorProjectionRequired || report.Coordinates.DeferredTopology != 1 {
			t.Fatal("completion classification changed the logical observation or its remaining work")
		}
	}
}
