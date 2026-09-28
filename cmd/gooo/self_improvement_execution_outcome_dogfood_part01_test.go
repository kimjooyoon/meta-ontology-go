package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodTypedChainClosesExecutionRequestWithOutcome(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve execution outcome fixture path")
	}
	fixture := filepath.Join(filepath.Dir(currentFile), "..", "..", "examples", "language-runtime-binding", "typed-chain.gooo")
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	const sourcePath = "examples/language-runtime-binding/typed-chain.gooo"
	plan, err := valueexecution.CompilePlan(sourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{"ProposeCandidate": 41})
	if err != nil {
		t.Fatal(err)
	}
	origin := dogfoodSelfImprovementOrigin(execution, "selected-candidate")
	receipt := valueexecution.ObserveExecutionOrigin(origin, execution)
	selection := valueexecution.SelfImprovementCandidateSelectionObservation{
		Status:         valueexecution.SelfImprovementCandidateSelectionStatusSelected,
		NonAuthorizing: true,
		Digest:         cache.HashBytes([]byte("selected-candidate-evidence")).String(),
	}
	request := valueexecution.ObserveSelfImprovementExecutionRequest(
		selection,
		valueexecution.SelfImprovementExecutionRequestContext{
			RequestID:              "typed-chain-execution-001",
			SourceURI:              sourcePath,
			EnvironmentDigest:      cache.HashBytes([]byte("typed-chain-environment")).String(),
			WorkloadIdentityDigest: cache.HashBytes([]byte("typed-chain-workload-identity")).String(),
			GatewayPolicyDigest:    cache.HashBytes([]byte("typed-chain-gateway-policy")).String(),
		},
	)
	outcome := valueexecution.ObserveSelfImprovementExecutionOutcome(request, receipt)
	if request.Status != valueexecution.SelfImprovementExecutionRequestStatusReady {
		t.Fatalf("request=%#v, want ready", request)
	}
	if outcome.Status != valueexecution.SelfImprovementExecutionOutcomeStatusCompleted {
		t.Fatalf("outcome=%#v, want completed", outcome)
	}
	if !outcome.NonAuthorizing || outcome.Digest == "" {
		t.Fatal("execution outcome must remain non-authorizing and digestable")
	}
}
