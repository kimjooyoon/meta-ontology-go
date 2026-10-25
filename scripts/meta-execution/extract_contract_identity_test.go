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
	if evidence.ContractSemanticDigest != contract.SemanticDigest {
		t.Fatal("contract meaning changed; review the decoder fixture's shape before adapting it")
	}
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
	if evidence.ContractSourceDigest == contract.SourceDigest || evidence.ContractSemanticDigest != contract.SemanticDigest {
		t.Fatal("fixture must isolate the changed source identity with unchanged semantics")
	}
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
