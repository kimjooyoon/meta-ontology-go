package bindingcoverage

import (
	"fmt"
	"reflect"
	"testing"

	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
)

const expectedProductionEquivalenceReceipt = "sha256:ab34297396618634dc46747a55ec5b386149df285308094005c84fd6b61eeb2c"

func TestProductionBindingCoverageEquivalence(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	if len(corpus.Cases) != 20 {
		t.Fatalf("case count=%d", len(corpus.Cases))
	}
	seen := make(map[string]bool, len(corpus.Cases))
	receipts := make([]productionReceipt, 0, len(corpus.Cases))
	byName := make(map[string]productionReceipt, len(corpus.Cases))
	mismatchCount := 0
	for _, row := range corpus.Cases {
		if seen[row.Name] {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s duplicate corpus name", row.Name)
			continue
		}
		seen[row.Name] = true
		oracle := Evaluate(row.Input)
		translated, translateErr := translateInput(row.Input)
		if translateErr != nil {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s translator UNKNOWN: %v", row.Name, translateErr)
			continue
		}
		productionOutput := production.Observe(translated)
		receipt, compareErr := compareProductionCase(row, oracle, translated, productionOutput)
		if compareErr != nil {
			mismatchCount++
			t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO case=%s: %v\noracle=%+v\nproduction=%+v", row.Name, compareErr, oracle.Vector, productionOutput)
		}
		receipts = append(receipts, receipt)
		byName[row.Name] = receipt
	}
	if len(seen) != len(corpus.Cases) {
		mismatchCount++
		t.Errorf("corpus consumption=%d/%d", len(seen), len(corpus.Cases))
	}
	if err := comparePermutation(byName); err != nil {
		mismatchCount++
		t.Errorf("PRODUCTION_EQUIVALENCE_COUNTEREXAMPLE_NO_GO permutation: %v", err)
	}
	digest, err := receiptDigest(receipts)
	if err != nil {
		t.Fatal(err)
	}
	if digest != expectedProductionEquivalenceReceipt {
		t.Fatalf("paired receipt digest=%s want=%s", digest, expectedProductionEquivalenceReceipt)
	}
	t.Logf("binding coverage paired receipt digest=%s cases=%d mismatch_count=%d", digest, len(corpus.Cases), mismatchCount)
	if mismatchCount != 0 {
		t.Fatalf("production equivalence mismatch_count=%d", mismatchCount)
	}
}

func compareProductionCase(row CorpusCase, oracle Result, input production.Input, output production.Output) (productionReceipt, error) {
	productionReason, err := expectedProductionReason(row.Name, oracle.Reason)
	if err != nil {
		return productionReceipt{}, err
	}
	inputBytes, inputDigest, err := productionInputReceipt(input, output)
	if err != nil {
		return productionReceipt{}, err
	}
	authorizationAbsent, err := authorizationFieldsAbsent(output)
	if err != nil {
		return productionReceipt{}, err
	}
	missingMatch, err := mapProductionIDs(output.MissingMatchBindingIDs)
	if err != nil {
		return productionReceipt{}, err
	}
	missingMismatch, err := mapProductionIDs(output.MissingMismatchBindingIDs)
	if err != nil {
		return productionReceipt{}, err
	}
	receipt := productionReceipt{
		CaseName: row.Name, Decision: string(output.Decision), Reason: string(output.Reason),
		RequiredBindingCount: output.RequiredBindingCount, PartitionCount: output.PartitionCount,
		EndpointReferenceCount: output.EndpointReferenceCount, WorkUnits: output.DeterministicWorkUnits,
		MissingMatch: missingMatch, MissingMismatch: missingMismatch,
		ProductionInputBytes: inputBytes, ProductionInputDigest: inputDigest,
		ProductionOutputDigest: output.CanonicalDigest, AuthorizationFieldsGone: authorizationAbsent,
	}
	checks := []string{}
	if output.Decision != production.Decision(oracle.Decision) {
		checks = append(checks, fmt.Sprintf("decision=%s want=%s", output.Decision, oracle.Decision))
	}
	if output.Reason != productionReason {
		checks = append(checks, fmt.Sprintf("reason=%s want=%s", output.Reason, productionReason))
	}
	checks = append(checks, compareCounts(oracle, output)...)
	if !reflect.DeepEqual(missingMatch, oracle.MissingMatch) {
		checks = append(checks, fmt.Sprintf("missing_match=%v want=%v", missingMatch, oracle.MissingMatch))
	}
	if !reflect.DeepEqual(missingMismatch, oracle.MissingMismatch) {
		checks = append(checks, fmt.Sprintf("missing_mismatch=%v want=%v", missingMismatch, oracle.MissingMismatch))
	}
	if output.SchemaVersion != input.SchemaVersion || output.ContractID != input.ContractID || output.SnapshotDigest != input.SnapshotDigest || output.ExpectedSnapshotDigest != input.ExpectedSnapshotDigest {
		checks = append(checks, "production provenance fields do not echo translated input")
	}
	if !authorizationAbsent {
		checks = append(checks, "authorization field present")
	}
	if len(checks) != 0 {
		return receipt, fmt.Errorf("%v", checks)
	}
	return receipt, nil
}

func compareCounts(oracle Result, output production.Output) []string {
	checks := []string{}
	if uint64(oracle.RequiredBindingCount) != output.RequiredBindingCount {
		checks = append(checks, fmt.Sprintf("required_binding_count=%d want=%d", output.RequiredBindingCount, oracle.RequiredBindingCount))
	}
	if uint64(oracle.PartitionCount) != output.PartitionCount {
		checks = append(checks, fmt.Sprintf("partition_count=%d want=%d", output.PartitionCount, oracle.PartitionCount))
	}
	if uint64(oracle.EndpointReferenceCount) != output.EndpointReferenceCount {
		checks = append(checks, fmt.Sprintf("endpoint_reference_count=%d want=%d", output.EndpointReferenceCount, oracle.EndpointReferenceCount))
	}
	if uint64(oracle.WorkUnits) != output.DeterministicWorkUnits {
		checks = append(checks, fmt.Sprintf("work_units=%d want=%d", output.DeterministicWorkUnits, oracle.WorkUnits))
	}
	return checks
}

func comparePermutation(receipts map[string]productionReceipt) error {
	canonical, ok := receipts["complete-two-bindings"]
	if !ok {
		return fmt.Errorf("missing complete-two-bindings receipt")
	}
	permuted, ok := receipts["permuted-complete-two-bindings"]
	if !ok {
		return fmt.Errorf("missing permuted-complete-two-bindings receipt")
	}
	if canonical.ProductionInputBytes != permuted.ProductionInputBytes || canonical.ProductionInputDigest != permuted.ProductionInputDigest || canonical.ProductionOutputDigest != permuted.ProductionOutputDigest {
		return fmt.Errorf("canonical production permutation changed: canonical=%+v permuted=%+v", canonical, permuted)
	}
	return nil
}
