package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodTypedChainRetainsRegressionCounterexample(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve regression dogfood fixture path")
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
	observation := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin,
		candidateOrigin,
		dogfoodSelfImprovementContext(baselinePlan),
		dogfoodSelfImprovementContext(candidatePlan),
		valueexecution.ObserveExecutionOrigin(baselineOrigin, baselineExecution),
		valueexecution.ObserveExecutionOrigin(candidateOrigin, candidateExecution),
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(candidateExecution, "regressed-metric"),
	)
	if observation.Status != valueexecution.SelfImprovementStatusRegressed {
		t.Fatalf("observation=%#v, want regressed", observation)
	}
	ledger := valueexecution.NewSelfImprovementEvidenceLedger(1).Append(observation)
	promotion := valueexecution.ObserveSelfImprovementPromotion(ledger)
	if promotion.Status != valueexecution.SelfImprovementPromotionStatusBlocked {
		t.Fatalf("promotion=%#v, want blocked", promotion)
	}
	if !promotion.NonAuthorizing || promotion.Digest == "" {
		t.Fatal("regression evidence must remain non-authorizing and digestable")
	}
}
