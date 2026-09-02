package proposalpredecessor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const routeIdentityCaseDenominator = 5

func TestRouteIdentityContractCases(t *testing.T) {
	predecessor := strings.Repeat("a", 40)
	current := strings.Repeat("b", 40)
	cases := []struct {
		id                 string
		collection         Collection
		decision           string
		observation        string
		resolution         string
		reason             string
		unknownClass       string
	}{
		{
			id: "normal-route-bound-closed",
			collection: Collection{RequestedRoute: RouteDev, ObservedRuns: 2, ExactRuns: 1,
				OtherRouteRuns: 1, ObservedJobs: 5, ExactJobs: 1, ObservedArtifacts: 4,
				ExactArtifacts: 1, Candidates: []Candidate{readyCandidate(predecessor, 1)}},
			decision: DecisionClosed, observation: DecisionClosed, resolution: ResolutionExact,
			reason: ReasonSelected,
		},
		{
			id: "other-route-only-unknown",
			collection: Collection{RequestedRoute: RouteDev, ObservedRuns: 1, OtherRouteRuns: 1},
			decision: ResolutionFailClosed, observation: DecisionUnknown, resolution: ResolutionLower,
			reason: ReasonNotFound, unknownClass: UnknownClassMissing,
		},
		{
			id: "missing-route-identity-unknown",
			collection: Collection{RequestedRoute: RouteDev, ObservedRuns: 1, RouteUnknownRuns: 1, Unresolved: 1},
			decision: ResolutionFailClosed, observation: DecisionUnknown, resolution: ResolutionLower,
			reason: ReasonRouteUnknown, unknownClass: UnknownClassRoute,
		},
		{
			id: "duplicate-route-candidates-unknown",
			collection: Collection{RequestedRoute: RouteDev, ObservedRuns: 2, ExactRuns: 2,
				ObservedJobs: 10, ExactJobs: 2, ObservedArtifacts: 8, ExactArtifacts: 2,
				Candidates: []Candidate{readyCandidate(predecessor, 1), readyCandidate(predecessor, 2)}},
			decision: ResolutionFailClosed, observation: DecisionUnknown, resolution: ResolutionLower,
			reason: ReasonAmbiguous, unknownClass: UnknownClassAmbiguous,
		},
		{
			id: "contradictory-success-refuted",
			collection: Collection{RequestedRoute: RouteDev, ObservedRuns: 1, ExactRuns: 1,
				Contradictions: 1, FailureReason: ReasonRouteContradiction},
			decision: ResolutionFailClosed, observation: DecisionRefuted, resolution: ResolutionExact,
			reason: ReasonRouteContradiction,
		},
	}
	if len(cases) != routeIdentityCaseDenominator {
		t.Fatalf("route identity denominator = %d, want %d", len(cases), routeIdentityCaseDenominator)
	}

	for _, test := range cases {
		t.Run(test.id, func(t *testing.T) {
			report, _, err := Select("owner/repository", current, predecessor, test.collection)
			if test.observation == DecisionClosed && err != nil {
				t.Fatalf("closed case error: %v", err)
			}
			if test.observation != DecisionClosed && err == nil {
				t.Fatal("non-closed case unexpectedly selected")
			}
			if report.Decision != test.decision || report.ObservationDecision != test.observation ||
				report.ObservationResolution != test.resolution || report.Reason != test.reason {
				t.Fatalf("report = %+v", report)
			}
			if test.unknownClass == "" {
				if report.Unknown != nil {
					t.Fatalf("unexpected unknown = %+v", report.Unknown)
				}
				return
			}
			if report.Unknown == nil || report.Unknown.UnknownClass != test.unknownClass ||
				report.Unknown.Stage == "" || report.Unknown.Step == "" || report.Unknown.Reason == "" ||
				report.Unknown.NextOperation == "" || report.Unknown.BlockedBy == nil {
				t.Fatalf("unknown evidence = %+v", report.Unknown)
			}
		})
	}
}

func TestCollectBindsSameSHAObservationToRequestedRoute(t *testing.T) {
	predecessor := strings.Repeat("a", 40)
	cases := []struct {
		name              string
		runs              []githubRun
		exactRuns         int
		otherRouteRuns    int
		routeUnknownRuns  int
		unresolved        int
	}{
		{
			name: "other route is not a candidate",
			runs: []githubRun{{ID: 1, RunAttempt: 1, HeadSHA: predecessor, HeadBranch: RouteMain, Event: "push", Status: "completed", Conclusion: "failure", Name: workflowName}},
			otherRouteRuns: 1,
		},
		{
			name:             "missing route is unknown",
			runs:             []githubRun{{ID: 2, RunAttempt: 1, HeadSHA: predecessor, Event: "push", Status: "completed", Conclusion: "failure", Name: workflowName}},
			routeUnknownRuns: 1,
		unresolved:       1,
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				if strings.Contains(request.URL.Path, "/actions/workflows/metric-counterfactual.yml/runs") {
					_ = json.NewEncoder(writer).Encode(runsEnvelope{TotalCount: len(test.runs), WorkflowRuns: test.runs})
					return
				}
				t.Fatalf("unexpected API request: %s", request.URL.Path)
			}))
			defer server.Close()

			collection, err := Collect(context.Background(), server.Client(), server.URL, "token", "owner/repository", predecessor, RouteDev)
			if err != nil {
				t.Fatal(err)
			}
			if collection.ExactRuns != test.exactRuns || collection.OtherRouteRuns != test.otherRouteRuns ||
				collection.RouteUnknownRuns != test.routeUnknownRuns || collection.Unresolved != test.unresolved {
				t.Fatalf("collection = %+v", collection)
			}
		})
	}
}
