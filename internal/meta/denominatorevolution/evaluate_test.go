package denominatorevolution

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSourceDerivedContractProducesExactReadOnlyReport(t *testing.T) {
	report := Evaluate(Input{Contract: testContract(t), HeadSHA: "0123456789012345678901234567890123456789", Source: testSource(t)})
	t.Logf("decision=%s resolution=%s reason=%s projection=%+v summary=%+v cases=%+v", report.Decision, report.Resolution, report.Reason, report.SourceProjection, report.Summary, report.Cases)
	if err := Validate(report); err != nil {
		t.Fatal(err)
	}
	if report.Summary.FixedDenominatorNumerator != DenominatorSize || report.Summary.FixedDenominatorDenominator != DenominatorSize {
		t.Fatalf("fixed denominator = %d/%d, want 5/5", report.Summary.FixedDenominatorNumerator, report.Summary.FixedDenominatorDenominator)
	}
	if report.Summary.LegalAdvanceNumerator != 1 || report.Summary.UnauthorizedRejectionNumerator != 1 || report.Summary.UnknownPredecessorNumerator != 1 {
		t.Fatalf("case summary = %+v", report.Summary)
	}
	if !guardrailsConform(report.Summary.Guardrails) || report.RepositoryWrites != 0 || report.MutationAuthority {
		t.Fatalf("unsafe summary = %+v", report.Summary)
	}
	if report.Cases[0].Receipt == nil || !guardrailsConform(report.Cases[0].Receipt.Guardrails) {
		t.Fatalf("receipt guardrails = %+v", report.Cases[0].Receipt)
	}
}

func TestContractDriftFailsClosed(t *testing.T) {
	contract := testContract(t)
	contract.Policy.NoAggregateEstimates = false
	report := Evaluate(Input{Contract: contract, HeadSHA: "0123456789012345678901234567890123456789", Source: testSource(t)})
	if report.Decision != "FAIL_CLOSED" || report.Reason != "DENOMINATOR_EVOLUTION_CONTRACT_DRIFT" {
		t.Fatalf("report = %+v", report)
	}
}

func testContract(t *testing.T) Contract {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "../../../examples/denominator-evolution/contract.json"))
	if err != nil {
		t.Fatal(err)
	}
	contract, err := DecodeContract(raw)
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func testSource(t *testing.T) []byte {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	raw, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "../../../examples/denominator-evolution/main.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
