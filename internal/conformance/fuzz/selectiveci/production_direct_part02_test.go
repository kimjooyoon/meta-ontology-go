package selectiveci

import (
	"fmt"
	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"reflect"
	"slices"
	"sort"
)

func directCorpusCase(corpus Corpus) (Case, error) {
	var result Case
	count := 0
	for _, fixture := range corpus.Cases {
		if fixture.Name == "direct" {
			result, count = fixture, count+1
		}
	}
	if count != 1 {
		return Case{}, fmt.Errorf("expected one direct case, found %d", count)
	}
	return result, nil
}
func normalizeOracle(result Result) directVector {
	return directVector{Decision: string(result.Decision), ReasonClass: string(result.Reason), PartialIDs: append([]string(nil), result.CommandIDs...), SelectedCommandIDs: append([]string(nil), result.CommandIDs...), Argv: result.Argv, ChangedIDs: append([]string(nil), result.CommandIDs...), CPUWorkUnits: result.CPUUnits, MemoryCeiling: result.MemoryCeiling, ProvenancePathCount: result.PathCount}
}
func sharedContractMatch(oracle, production directVector) bool {
	if oracle.Decision != production.Decision || oracle.ReasonClass != production.ReasonClass {
		return false
	}
	if !reflect.DeepEqual(oracle.SelectedCommandIDs, production.SelectedCommandIDs) || !reflect.DeepEqual(oracle.SelectedGuardIDs, production.SelectedGuardIDs) || !reflect.DeepEqual(oracle.Argv, production.Argv) {
		return false
	}
	if !reflect.DeepEqual(oracle.ChangedIDs, production.ChangedIDs) || oracle.CPUWorkUnits != production.CPUWorkUnits || oracle.MemoryCeiling != production.MemoryCeiling || oracle.ProvenancePathCount != production.ProvenancePathCount {
		return false
	}
	return noFallbackPartials(oracle) && noFallbackPartials(production)
}
func noFallbackPartials(vector directVector) bool {
	if vector.Decision != string(FullSuiteFallback) {
		return true
	}
	return len(vector.PartialIDs) == 0 && len(vector.SelectedCommandIDs) == 0 && len(vector.SelectedGuardIDs) == 0 && len(vector.SelectedWorkIDs) == 0
}
func verifyProductionWorkIDs(fixture directFixture, result productionsci.PlanResult) bool {
	expected := make([]string, 0, len(result.SelectedCommandIDs)+len(result.SelectedGuardCommandIDs))
	for _, commandID := range result.SelectedCommandIDs {
		obligationID := directObligationFor(fixture.Input.Registry.Obligations, commandID)
		if obligationID == "" {
			return false
		}
		expected = append(expected, directWorkID(result.HeadSnapshotDigest, obligationID, commandID, fixture.Input.Registry.PolicyDigest))
	}
	for _, commandID := range result.SelectedGuardCommandIDs {
		expected = append(expected, directWorkID(result.HeadSnapshotDigest, "guard/"+commandID, commandID, fixture.Input.Registry.PolicyDigest))
	}
	sort.Strings(expected)
	actual := append([]string(nil), result.SelectedWorkIDs...)
	sort.Strings(actual)
	return reflect.DeepEqual(expected, actual)
}
func directObligationFor(bindings []productionsci.ObligationBinding, commandID string) string {
	for _, binding := range bindings {
		if slices.Contains(binding.CommandIDs, commandID) {
			return binding.ID
		}
	}
	return ""
}
