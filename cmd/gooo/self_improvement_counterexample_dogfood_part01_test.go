package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodTypedChainRetainsReplayableCounterexample(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve counterexample dogfood fixture path")
	}
	fixture := filepath.Join(filepath.Dir(currentFile), "..", "..", "examples", "language-runtime-binding", "typed-chain.gooo")
	source, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	const sourcePath = "examples/language-runtime-binding/typed-chain.gooo"
	baselinePlan, err := valueexecution.CompilePlan(sourcePath, source)
	if err != nil {
		t.Fatal(err)
	}
	baselineExecution, err := baselinePlan.Execute(map[string]int64{"ProposeCandidate": 41})
	if err != nil {
		t.Fatal(err)
	}
	oldDeclaration := []byte(`activity CommitCandidate(Integer) -> Integer computes "int.add:1"`)
	regressedDeclaration := []byte(`activity CommitCandidate(Integer) -> Integer computes "int.add:0"`)
	candidateSource := bytes.Replace(source, oldDeclaration, regressedDeclaration, 1)
	if bytes.Equal(candidateSource, source) {
		t.Fatal("regressed candidate source was not changed")
	}
	candidatePlan, err := valueexecution.CompilePlan(sourcePath, candidateSource)
	if err != nil {
		t.Fatal(err)
	}
	candidateExecution, err := candidatePlan.Execute(map[string]int64{"ProposeCandidate": 41})
	if err != nil {
		t.Fatal(err)
	}
	baselineOrigin := dogfoodSelfImprovementOrigin(baselineExecution, "baseline")
	candidateOrigin := dogfoodSelfImprovementOrigin(candidateExecution, "regressed-candidate")
	beforeReceipt := valueexecution.ObserveExecutionOrigin(baselineOrigin, baselineExecution)
	afterReceipt := valueexecution.ObserveExecutionOrigin(candidateOrigin, candidateExecution)
	observation := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin,
		candidateOrigin,
		dogfoodSelfImprovementContext(baselinePlan),
		dogfoodSelfImprovementContext(candidatePlan),
		beforeReceipt,
		afterReceipt,
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(candidateExecution, "regressed-metric"),
	)
	counterexample := valueexecution.ObserveSelfImprovementCounterexample(
		sourcePath,
		cache.HashBytes([]byte("ProposeCandidate=41")).String(),
		observation,
		beforeReceipt,
		afterReceipt,
	)
	if counterexample.Status != valueexecution.SelfImprovementCounterexampleStatusRetained {
		t.Fatalf("counterexample=%#v, want retained", counterexample)
	}
	if counterexample.SourceURI != sourcePath || counterexample.InputDigest == "" {
		t.Fatalf("counterexample=%#v, want source and input provenance", counterexample)
	}
	if !counterexample.NonAuthorizing || counterexample.Digest == "" {
		t.Fatal("counterexample must remain non-authorizing and digestable")
	}
}
