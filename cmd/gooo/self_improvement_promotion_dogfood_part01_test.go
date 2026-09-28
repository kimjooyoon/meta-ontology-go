package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodTypedChainBindsPromotionToExecutionReceipt(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve promotion dogfood fixture path")
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
	newDeclaration := []byte(`activity CommitCandidate(Integer) -> Integer computes "int.add:2"`)
	candidateSource := bytes.Replace(source, oldDeclaration, newDeclaration, 1)
	if bytes.Equal(candidateSource, source) {
		t.Fatal("candidate source was not changed")
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
	candidateOrigin := dogfoodSelfImprovementOrigin(candidateExecution, "candidate")
	observation := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin,
		candidateOrigin,
		dogfoodSelfImprovementContext(baselinePlan),
		dogfoodSelfImprovementContext(candidatePlan),
		valueexecution.ObserveExecutionOrigin(baselineOrigin, baselineExecution),
		valueexecution.ObserveExecutionOrigin(candidateOrigin, candidateExecution),
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(candidateExecution, "candidate-metric"),
	)
	ledger := valueexecution.NewSelfImprovementEvidenceLedger(1).Append(observation)
	promotion := valueexecution.ObserveSelfImprovementPromotion(ledger)
	binding := valueexecution.ObserveSelfImprovementPromotionBinding(
		promotion,
		valueexecution.ObserveExecutionOrigin(candidateOrigin, candidateExecution),
	)
	if promotion.Status != valueexecution.SelfImprovementPromotionStatusEligible {
		t.Fatalf("promotion=%#v, want eligible", promotion)
	}
	if binding.Status != valueexecution.SelfImprovementPromotionBindingStatusComplete {
		t.Fatalf("binding=%#v, want complete", binding)
	}
	if !binding.NonAuthorizing || binding.Digest == "" {
		t.Fatal("dogfood binding must remain non-authorizing and digestable")
	}
}
