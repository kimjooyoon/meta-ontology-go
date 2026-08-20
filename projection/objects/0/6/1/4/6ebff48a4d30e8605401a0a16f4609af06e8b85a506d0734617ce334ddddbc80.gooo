package selectiveci

import (
	"encoding/json"
	"testing"
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
