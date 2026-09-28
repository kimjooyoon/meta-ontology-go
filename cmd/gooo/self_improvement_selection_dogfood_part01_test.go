package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodCandidateSelectionBindsImprovementAndReplay(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve candidate selection fixture path")
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
	baselineOrigin := dogfoodSelfImprovementOrigin(baselineExecution, "baseline")
	baselineReceipt := valueexecution.ObserveExecutionOrigin(baselineOrigin, baselineExecution)

	improvedPlan, improvedExecution := dogfoodSelectionCandidate(t, sourcePath, source, 2)
	improvedOrigin := dogfoodSelfImprovementOrigin(improvedExecution, "improved-candidate")
	improvedReceipt := valueexecution.ObserveExecutionOrigin(improvedOrigin, improvedExecution)
	improvement := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin,
		improvedOrigin,
		dogfoodSelfImprovementContext(baselinePlan),
		dogfoodSelfImprovementContext(improvedPlan),
		baselineReceipt,
		improvedReceipt,
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(improvedExecution, "improved-metric"),
	)
	promotion := valueexecution.ObserveSelfImprovementPromotion(
		valueexecution.NewSelfImprovementEvidenceLedger(1).Append(improvement),
	)

	regressedPlan, regressedExecution := dogfoodSelectionCandidate(t, sourcePath, source, 0)
	regressedOrigin := dogfoodSelfImprovementOrigin(regressedExecution, "regressed-candidate")
	regressedReceipt := valueexecution.ObserveExecutionOrigin(regressedOrigin, regressedExecution)
	regression := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin,
		regressedOrigin,
		dogfoodSelfImprovementContext(baselinePlan),
		dogfoodSelfImprovementContext(regressedPlan),
		baselineReceipt,
		regressedReceipt,
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(regressedExecution, "regressed-metric"),
	)
	counterexample := valueexecution.ObserveSelfImprovementCounterexample(
		sourcePath,
		cache.HashBytes([]byte("ProposeCandidate=41")).String(),
		regression,
		baselineReceipt,
		regressedReceipt,
	)
	replay := valueexecution.ObserveSelfImprovementCounterexampleReplay(counterexample, counterexample)
	selection := valueexecution.ObserveSelfImprovementCandidateSelection(promotion, replay)
	if promotion.Status != valueexecution.SelfImprovementPromotionStatusEligible {
		t.Fatalf("promotion=%#v, want eligible", promotion)
	}
	if replay.Status != valueexecution.SelfImprovementCounterexampleReplayStatusMatched {
		t.Fatalf("replay=%#v, want matched", replay)
	}
	if selection.Status != valueexecution.SelfImprovementCandidateSelectionStatusSelected {
		t.Fatalf("selection=%#v, want selected", selection)
	}
	if !selection.NonAuthorizing || selection.Digest == "" {
		t.Fatal("selection must remain non-authorizing and digestable")
	}
}

func dogfoodSelectionCandidate(t *testing.T, sourcePath string, source []byte, delta int64) (valueexecution.Plan, valueexecution.Execution) {
	t.Helper()
	oldDeclaration := []byte(`activity CommitCandidate(Integer) -> Integer computes "int.add:1"`)
	newDeclaration := []byte(fmt.Sprintf(`activity CommitCandidate(Integer) -> Integer computes "int.add:%d"`, delta))
	candidateSource := bytes.Replace(source, oldDeclaration, newDeclaration, 1)
	if bytes.Equal(candidateSource, source) {
		t.Fatalf("candidate delta %d did not change source", delta)
	}
	plan, err := valueexecution.CompilePlan(sourcePath, candidateSource)
	if err != nil {
		t.Fatal(err)
	}
	execution, err := plan.Execute(map[string]int64{"ProposeCandidate": 41})
	if err != nil {
		t.Fatal(err)
	}
	return plan, execution
}
