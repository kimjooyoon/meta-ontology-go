package selectiveci

import (
	"encoding/json"
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"sort"
	"testing"
)

// directWorkIDEvidence preserves the first counterexample: the oracle has no public work-ID field, while production selected one.
func assertOriginalWorkIDGapEvidence(t *testing.T, oracle, production directVector) {
	t.Helper()
	if len(oracle.SelectedWorkIDs) != 0 || len(production.SelectedWorkIDs) == 0 {
		t.Fatalf("named direct work-ID gap evidence changed: oracle=%#v production=%#v", oracle.SelectedWorkIDs, production.SelectedWorkIDs)
	}
	oracleJSON, _ := json.Marshal(oracle)
	productionJSON, _ := json.Marshal(production)
	t.Logf("directWorkIDEvidence oracle=%s production=%s", oracleJSON, productionJSON)
}
func normalizeProduction(fixture directFixture, result productionsci.PlanResult) directVector {
	commands := map[string]productionsci.Command{}
	for _, command := range append(fixture.Input.Registry.Commands, fixture.Input.Registry.GlobalGuardCommands...) {
		commands[command.ID] = command
	}
	vector := directVector{Decision: string(result.Status), ReasonClass: directReasonClass(result), SelectedCommandIDs: mapIDs(result.SelectedCommandIDs, fixture.ProdToOracle), SelectedGuardIDs: append([]string(nil), result.SelectedGuardCommandIDs...), SelectedWorkIDs: append([]string(nil), result.SelectedWorkIDs...), ChangedIDs: mapIDs(result.ChangedSemanticIDs, fixture.ProdToOracle), Argv: mapArgv(result, fixture, commands), ProvenancePathCount: len(result.ProvenancePathIDs)}
	vector.PartialIDs = append(vector.PartialIDs, vector.SelectedCommandIDs...)
	vector.PartialIDs = append(vector.PartialIDs, vector.SelectedGuardIDs...)
	vector.PartialIDs = append(vector.PartialIDs, vector.SelectedWorkIDs...)
	for _, id := range append(result.SelectedCommandIDs, result.SelectedGuardCommandIDs...) {
		vector.CPUWorkUnits += commands[id].CPUWorkUnits
		vector.MemoryCeiling += commands[id].MemoryBytes
	}
	return vector
}
func directReasonClass(result productionsci.PlanResult) string {
	if result.Status == productionsci.StatusSelective && result.ReasonCode == productionsci.ReasonNone {
		return string(completeReason)
	}
	switch result.ReasonCode {
	case productionsci.ReasonUnknownPath:
		return string(unknownPathReason)
	case productionsci.ReasonInvalidArgv:
		return string(emptyArgvReason)
	case productionsci.ReasonAmbiguousPath:
		return string(ambiguousReason)
	case productionsci.ReasonMismatchedDigest:
		return string(blobMismatchReason)
	case productionsci.ReasonDanglingReference:
		return string(danglingCmdReason)
	case productionsci.ReasonDuplicateID:
		return string(duplicateIDReason)
	default:
		return "UNEXPECTED:" + result.ReasonCode
	}
}
func mapIDs(ids []string, reverse map[string]string) []string {
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		if mapped, ok := reverse[id]; ok {
			result = append(result, mapped)
		} else {
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result
}
