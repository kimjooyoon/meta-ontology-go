package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func TestDogfoodTypedChainProducesSelfImprovementLedgerEvidence(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve typed runtime chain fixture path")
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
	if baselineExecution.Results["CommitCandidate"].Value != 44 || candidateExecution.Results["CommitCandidate"].Value != 45 {
		t.Fatalf("baseline=%#v candidate=%#v, want source-backed result change", baselineExecution.Results["CommitCandidate"], candidateExecution.Results["CommitCandidate"])
	}

	baselineOrigin := dogfoodSelfImprovementOrigin(baselineExecution, "baseline")
	candidateOrigin := dogfoodSelfImprovementOrigin(candidateExecution, "candidate")
	baselineContext := dogfoodSelfImprovementContext(baselinePlan)
	candidateContext := dogfoodSelfImprovementContext(candidatePlan)
	observation := valueexecution.ObserveExecutedSelfImprovementWithContext(
		baselineOrigin, candidateOrigin, baselineContext, candidateContext,
		valueexecution.ObserveExecutionOrigin(baselineOrigin, baselineExecution),
		valueexecution.ObserveExecutionOrigin(candidateOrigin, candidateExecution),
		dogfoodSelfImprovementMetric(baselineExecution, "baseline-metric"),
		dogfoodSelfImprovementMetric(candidateExecution, "candidate-metric"),
	)
	if observation.Status != valueexecution.SelfImprovementStatusImproved || observation.Reason != "METRIC_IMPROVED" {
		t.Fatalf("observation=%#v, want source-backed improvement", observation)
	}
	ledger := valueexecution.NewSelfImprovementEvidenceLedger(1).Append(observation).Observe()
	if ledger.Decision != valueexecution.SelfImprovementLedgerDecisionPromotionEligible || ledger.Improved != 1 || ledger.Unknown != 0 {
		t.Fatalf("ledger=%#v, want one eligible measured improvement", ledger)
	}
}

func dogfoodSelfImprovementContext(plan valueexecution.Plan) valueexecution.ImprovementContextObservation {
	return valueexecution.ObserveImprovementEvaluationContext(valueexecution.EnvironmentObservation{
		Inputs: valueexecution.EnvironmentInputs{
			SourceDigest:        plan.SourceDigest,
			SemanticFingerprint: plan.SemanticFingerprint,
			ToolchainDigest:     "sha256:" + cache.HashBytes([]byte("go-toolchain")).String(),
			ContractDigest:      "sha256:" + cache.HashBytes([]byte("gooo-contract-v1")).String(),
			ModelDigest:         "sha256:" + cache.HashBytes([]byte("gooo-model-v1")).String(),
			SkillDigest:         "sha256:" + cache.HashBytes([]byte("gooo-skill-v1")).String(),
			GatewayPolicyDigest: "sha256:" + cache.HashBytes([]byte("gooo-gateway-policy-v1")).String(),
		},
	})
}

func dogfoodSelfImprovementOrigin(execution valueexecution.Execution, label string) provenance.OriginChainObservation {
	encoded, _ := json.Marshal(execution)
	return provenance.ObserveOriginChain(provenance.OriginChain{
		DeclarationURI:        "examples/language-runtime-binding/typed-chain.gooo",
		DeclarationSymbol:     "activity CommitCandidate",
		IRNode:                "activity:CommitCandidate",
		GeneratedURI:          "cmd/gooo/run_source_typed_chain_part01_test.go",
		GeneratedSymbol:       "TestRunSourceTypedRuntimeChainHasIndependentReplayEvidence",
		ReverseObservationURI: "internal/valueexecution/execution_origin_receipt_part01.go",
		ReverseObservation:    "execution.completed",
		MetricName:            "gooo.self-improvement.metric.v1",
		MetricValue:           fmt.Sprintf("%s=%d", label, execution.Results["CommitCandidate"].Value),
		EvidenceDigest:        "sha256:" + cache.HashBytes(encoded).String(),
	})
}

func dogfoodSelfImprovementMetric(execution valueexecution.Execution, label string) valueexecution.ImprovementMetric {
	encoded, _ := json.Marshal(execution.Results["CommitCandidate"])
	return valueexecution.ImprovementMetric{
		Name:           "commit_candidate_value",
		Value:          execution.Results["CommitCandidate"].Value,
		EvidenceDigest: "sha256:" + cache.HashBytes(append([]byte(label+":"), encoded...)).String(),
		LowerIsBetter:  false,
	}
}
