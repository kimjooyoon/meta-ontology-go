package selectiveci

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	productionsci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
)

type directVector struct {
	Decision            string              `json:"decision"`
	ReasonClass         string              `json:"reason_class"`
	PartialIDs          []string            `json:"partial_ids"`
	SelectedCommandIDs  []string            `json:"selected_command_ids"`
	SelectedGuardIDs    []string            `json:"selected_guard_ids"`
	SelectedWorkIDs     []string            `json:"selected_work_ids"`
	Argv                map[string][]string `json:"argv"`
	ChangedIDs          []string            `json:"changed_ids"`
	CPUWorkUnits        uint64              `json:"cpu_work_units"`
	MemoryCeiling       uint64              `json:"memory_ceiling"`
	ProvenancePathCount int                 `json:"provenance_path_count"`
}

type directReceipt struct {
	CaseID                      string       `json:"case_id"`
	SharedContractMatch         bool         `json:"shared_contract_match"`
	ProductionExtensionVerified bool         `json:"production_extension_verified"`
	OracleCoverage              string       `json:"oracle_coverage"`
	Equivalence                 string       `json:"equivalence"`
	Oracle                      directVector `json:"oracle"`
	Production                  directVector `json:"production"`
}

func TestProductionDirectCounterexample(t *testing.T) {
	corpus, err := LoadCorpus()
	if err != nil {
		t.Fatal(err)
	}
	fixtureCase, err := directCorpusCase(corpus)
	if err != nil {
		t.Fatal(err)
	}
	oracle := Evaluate(fixtureCase)
	fixture, err := translateDirect(fixtureCase)
	if err != nil {
		t.Fatalf("DIRECT_COUNTEREXAMPLE_NO_GO case=direct translator=%v", err)
	}
	production, err := runDirect(fixture)
	if err != nil {
		t.Fatalf("DIRECT_COUNTEREXAMPLE_NO_GO case=direct production=%v", err)
	}
	oracleVector := normalizeOracle(oracle)
	productionVector := normalizeProduction(fixture, production)
	sharedMatch := sharedContractMatch(oracleVector, productionVector)
	extensionVerified := verifyProductionWorkIDs(fixture, production)
	assertOriginalWorkIDGapEvidence(t, oracleVector, productionVector)
	receipt := directReceipt{CaseID: "direct", SharedContractMatch: sharedMatch, ProductionExtensionVerified: extensionVerified, OracleCoverage: "PARTIAL_WORK_ID_UNSPECIFIED", Equivalence: "UNKNOWN", Oracle: oracleVector, Production: productionVector}
	digest := directReceiptDigest(receipt)
	t.Logf("direct paired-vector receipt digest=%s", digest)
	if !sharedMatch || !extensionVerified || receipt.OracleCoverage != "PARTIAL_WORK_ID_UNSPECIFIED" || receipt.Equivalence != "UNKNOWN" {
		oracleJSON, _ := json.Marshal(oracleVector)
		productionJSON, _ := json.Marshal(productionVector)
		t.Fatalf("direct classification precondition failed receipt=%s oracle=%s production=%s", digest, oracleJSON, productionJSON)
	}
	t.Logf("direct classification oracle_coverage=PARTIAL_WORK_ID_UNSPECIFIED equivalence=UNKNOWN receipt=%s", digest)
}

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
		for _, candidate := range binding.CommandIDs {
			if candidate == commandID {
				return binding.ID
			}
		}
	}
	return ""
}

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

func mapArgv(result productionsci.PlanResult, fixture directFixture, commands map[string]productionsci.Command) map[string][]string {
	argv := make(map[string][]string)
	for _, id := range append(result.SelectedCommandIDs, result.SelectedGuardCommandIDs...) {
		key := id
		if mapped, ok := fixture.ProdToOracle[id]; ok {
			key = mapped
		}
		argv[key] = append([]string(nil), commands[id].Argv...)
	}
	return argv
}
