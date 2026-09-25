package proposalpredecessor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPendingCollectionPreservesLegacyQueryAndBindsFrontier(t *testing.T) {
	head := strings.Repeat("a", 40)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if !strings.HasSuffix(r.URL.Path, "/workflows/metric-counterfactual.yml/runs") {
			t.Errorf("unexpected API request: %s", r.URL)
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		runs := []githubRun{}
		if r.URL.Query().Get("status") == "" {
			runs = append(runs, githubRun{ID: 42, RunAttempt: 1, HeadSHA: head,
				HeadBranch: RouteDev, Event: "push", Status: "in_progress", Name: workflowName})
		}
		_ = json.NewEncoder(w).Encode(runsEnvelope{TotalCount: len(runs), WorkflowRuns: runs})
	}))
	defer server.Close()
	legacy, err := Collect(context.Background(), server.Client(), server.URL, "token", "owner/repo", head, RouteDev)
	if err != nil || legacy.ObservedRuns != 0 {
		t.Fatalf("legacy query changed: %+v %v", legacy, err)
	}
	collection, err := CollectPending(context.Background(), server.Client(), server.URL, "token", "owner/repo", head, RouteDev)
	if err != nil || collection.ObservedRuns != 1 || collection.Unresolved != 1 {
		t.Fatalf("pending collection: %+v %v", collection, err)
	}
	report, payload, err := SelectPending("owner/repo", strings.Repeat("b", 40), head, collection)
	if err == nil || len(payload) != 0 || !AwaitablePending(report) || requests != 2 {
		t.Fatalf("pending selection: %+v payload=%q requests=%d err=%v", report, payload, requests, err)
	}
	if report.Unknown.BlockedBy[0] != "workflow_run:42:attempt:1" ||
		report.Summary.ProofsTotal != 5 || report.Summary.ProofsPassed != 0 {
		t.Fatalf("frontier or proof denominator changed: %+v", report)
	}
}
