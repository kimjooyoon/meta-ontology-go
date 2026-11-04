package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	projectionextractor "github.com/kimjooyoon/meta-ontology-go/internal/meta/repositoryprojection/extractor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

// This is a synthetic decoder-shape fixture with the current admission identity.
// The historical artifact stays unchanged and is checked separately below.
// No execution or native proof is asserted by rebinding these test-only bytes.
func currentStrategySchemaFixture(t *testing.T) projectionextractor.StrategyEvidence {
	t.Helper()
	evidence := historicalStrategyEvidenceFixture(t)
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	assertHistoricalExtractionShape(t, evidence, contract)
	evidence.ContractSemanticDigest = contract.SemanticDigest
	evidence.ContractSourceDigest = contract.SourceDigest
	evidence.ContractInputEntity, evidence.ContractInputEntityID = contract.InputEntity, contract.InputEntityID
	evidence.ContractOutputEntity, evidence.ContractOutputEntityID = contract.OutputEntity, contract.OutputEntityID
	for index, expected := range contract.Obligations {
		obligation, stage := &evidence.ContractObligations[index], &evidence.ProofStages[index]
		obligation.InputEntity, obligation.InputEntityID = expected.InputEntity, expected.InputEntityID
		obligation.OutputEntity, obligation.OutputEntityID = expected.OutputEntity, expected.OutputEntityID
		stage.InputEntity, stage.InputEntityID = expected.InputEntity, expected.InputEntityID
		stage.OutputEntity, stage.OutputEntityID = expected.OutputEntity, expected.OutputEntityID
	}
	return evidence
}

func TestHistoricalStrategyIdentityIsNotCurrentAdmission(t *testing.T) {
	evidence := historicalStrategyEvidenceFixture(t)
	contract, err := generation.ExtractFunctionInputContractEvidence()
	if err != nil {
		t.Fatal(err)
	}
	if evidence.ContractSourceDigest == contract.SourceDigest {
		t.Fatal("historical fixture acquired the current source identity")
	}
	assertHistoricalExtractionShape(t, evidence, contract)
	assertHistoricalReportRejected(t, evidence)
}

func TestPreviousSourceWithCurrentShapeRetainsActionableMismatch(t *testing.T) {
	evidence := currentStrategySchemaFixture(t)
	evidence.ContractSourceDigest = historicalStrategyEvidenceFixture(t).ContractSourceDigest
	report := extractionReportWithStrategyEvidence([]projectionextractor.StrategyEvidence{evidence})
	payload, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	path := filepath.Join(root, "historical-report.json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := decodeExtractorReport(path, "head"); !errors.Is(err, errExtractorContractSourceMismatch) {
		t.Fatalf("historical source did not retain the actionable contract mismatch: %v", err)
	}
	assertContractMismatchReachesDrivers(t, root, "historical-report.json")
}

func assertHistoricalReportRejected(t *testing.T, evidence projectionextractor.StrategyEvidence) {
	t.Helper()
	report := extractionReportWithStrategyEvidence([]projectionextractor.StrategyEvidence{evidence})
	path := writeMapConstructionConsumerReport(t, t.TempDir(), report)
	if _, _, err := decodeExtractorReport(path, "head"); err == nil {
		t.Fatal("historical contract was admitted as current execution evidence")
	}
}

// Compare the extraction subgraph to the retained original fixture before
// constructing decoder-only current identities. The added Go split subgraph
// changes the whole contract digest; these six extraction stages retain shape.
func assertHistoricalExtractionShape(t *testing.T, old projectionextractor.StrategyEvidence,
	current generation.OperationInputContractEvidence) {
	t.Helper()
	if old.ContractActivity != current.Activity || old.ContractInputSubjectKind != string(current.InputSubjectKind) ||
		len(old.ContractObligations) != 6 || len(current.Obligations) != 6 || len(old.ProofStages) != 6 ||
		current.InputEntityID != "gooo://meta-operation-input-contract/entity/function-input" ||
		current.OutputEntityID != "gooo://meta-operation-input-contract/entity/result" {
		t.Fatal("current extraction shape differs from the retained six-stage fixture")
	}
	roles := []string{"function-input", "return-shape-obligation", "control-flow-obligation", "free-bindings-obligation",
		"callee-effects-obligation", "rendered-capacity-obligation", "projected-conformance-obligation"}
	for index, expected := range old.ContractObligations {
		got := current.Obligations[index]
		if got.Name != expected.Name || got.Activity != expected.Activity || !got.UsedInputFact || !got.GeneratedOutputFact ||
			!expected.UsedInputFact || !expected.GeneratedOutputFact ||
			got.InputEntityID != "gooo://meta-operation-input-contract/entity/"+roles[index] ||
			got.OutputEntityID != "gooo://meta-operation-input-contract/entity/"+roles[index+1] {
			t.Fatalf("extraction stage %d differs from the historical shape", index)
		}
	}
}

func TestPreparationProgressRequiresCurrentEntityIdentity(t *testing.T) {
	evidence := intermediatePreparationEvidenceFixture(t)
	assertExtractorReportMapAccepted(t, extractionReportMapWithStrategyEvidence(t, evidence))
	evidence.PreparationProgress.InputEntityID = "wrong-identity-with-same-label"
	report := extractionReportWithStrategyEvidence([]projectionextractor.StrategyEvidence{evidence})
	path := writeMapConstructionConsumerReport(t, t.TempDir(), report)
	if _, _, err := decodeExtractorReport(path, "head"); err == nil {
		t.Fatal("progress from a different entity identity was accepted")
	}
}

func assertContractMismatchReachesDrivers(t *testing.T, root, reportName string) {
	t.Helper()
	plan, action := generation.Plan{HeadSHA: "head"}, generation.Action{}
	observation := generation.ProcessObservation{ExitCode: 0, StdoutBytes: 2}
	materialized, evaluated := evaluateExtractMaterialization(root, nil, nil,
		processResult{Observation: observation}, plan, action, sourcepolicy.SourceSubject{}, reportName, metaExecutionTrace{}, "")
	if !reflect.DeepEqual(materialized, operationMaterialization{Executor: observation}) {
		t.Fatal("rejected historical report produced admitted materialization")
	}
	failed := failedExtractionError(root, reportName, plan, action)
	for _, failure := range []*operationError{evaluated, failed} {
		if failure == nil || failure.reason != extractionContractSourceMismatch ||
			failure.stage != "evaluate-operation" || failure.step != "decode-function-extraction-report" ||
			failure.class != "KNOWN_CONTRADICTION" || failure.next != "restore-extraction-plan" {
			t.Fatalf("driver hid the source mismatch or next operation: %+v", failure)
		}
	}
}
