package bindingcoverage

import (
	"fmt"
	production "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci/bindingcoverage"
	"reflect"
)

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
