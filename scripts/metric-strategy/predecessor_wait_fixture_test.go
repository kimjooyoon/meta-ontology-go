package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	predecessor "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

func pendingObservationFixture(t *testing.T) predecessor.Report {
	t.Helper()
	head := strings.Repeat("a", 40)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, `{"total_count":1,"workflow_runs":[{"id":42,"run_attempt":1,"head_sha":%q,"head_branch":"dev","event":"push","status":"in_progress","conclusion":null,"name":"Metric counterfactual conformance"}]}`, head)
	}))
	defer server.Close()
	collection, err := predecessor.CollectPending(context.Background(), server.Client(), server.URL, "token", "owner/repo", head, predecessor.RouteDev)
	if err != nil {
		t.Fatal(err)
	}
	report, _, err := predecessor.SelectPending("owner/repo", strings.Repeat("b", 40), head, collection)
	if err == nil || !predecessor.AwaitablePending(report) {
		t.Fatalf("invalid pending fixture: %+v %v", report, err)
	}
	return report
}

func TestPredecessorObservationRejectsInvalidBudgets(t *testing.T) {
	for _, attempts := range []int{0, maximumPredecessorObservations + 1} {
		_, _, err := awaitProposalPredecessor(context.Background(), attempts, nil, nil, nil)
		if err == nil {
			t.Fatalf("accepted invalid attempt limit %d", attempts)
		}
	}
}
