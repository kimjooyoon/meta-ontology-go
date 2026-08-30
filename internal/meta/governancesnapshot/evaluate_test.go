package governancesnapshot

import (
	"encoding/json"
	"os"
	"testing"
)

func contractFixture(t *testing.T) Contract {
	t.Helper()
	raw, err := os.ReadFile("../../../examples/live-governance-snapshot/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract Contract
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContract(contract); err != nil {
		t.Fatal(err)
	}
	return contract
}

func graphFixture(contract Contract) RawGraph {
	graph := RawGraph{SchemaVersion: GraphSchema, SourceDigest: "sha256:source", GraphHash: "sha256:graph"}
	for _, spec := range contract.Cells {
		graph.Nodes = append(graph.Nodes,
			GraphNode{ID: spec.InputID, Kind: "Entity", Name: spec.InputID},
			GraphNode{ID: spec.OutputID, Kind: "Entity", Name: spec.OutputID},
			GraphNode{ID: "activity:" + spec.Activity, Kind: "Activity", Name: spec.Activity})
		graph.Relations = append(graph.Relations,
			GraphRelation{Subject: "activity:" + spec.Activity, Predicate: "used", Object: spec.InputID},
			GraphRelation{Subject: spec.OutputID, Predicate: "wasGeneratedBy", Object: "activity:" + spec.Activity})
	}
	return graph
}

func TestCurrentDriftIsRefuted(t *testing.T) {
	contract := contractFixture(t)
	graph := graphFixture(contract)
	report := Evaluate(fixtureSnapshot(contract, "drift"), contract, graph)
	if err := ValidateReport(report, contract, graph, "fixture-head"); err != nil {
		t.Fatal(err)
	}
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact || report.Reason != "DEV_LIVE_PROTECTION_DRIFT" ||
		report.Summary != (Summary{CellsTotal: 12, ClosedCells: 10, RefutedCells: 2, FoundationCells: 4, CoherenceCells: 4, RegressionCells: 4}) ||
		report.SettingsHealthy || report.PromotionAuthorized {
		t.Fatalf("drift report = %#v", report)
	}
}

func TestCanonicalCasesPreserveUnknownFrontiers(t *testing.T) {
	contract := contractFixture(t)
	graph := graphFixture(contract)
	report := Evaluate(fixtureSnapshot(contract, "normal"), contract, graph)
	if err := ValidateReport(report, contract, graph, "fixture-head"); err != nil {
		t.Fatal(err)
	}
	if len(report.Cases) != 6 || report.Cases[2].Unknown == nil || report.Cases[3].Unknown == nil {
		t.Fatalf("cases = %#v", report.Cases)
	}
	if report.Cases[2].Unknown.UnknownClass != "DIRECT_MISSING" || len(report.Cases[2].Unknown.BlockedBy) != 0 {
		t.Fatalf("missing case = %#v", report.Cases[2])
	}
	if report.Cases[3].Unknown.UnknownClass != "DEPENDENCY_BLOCKED" || len(report.Cases[3].Unknown.BlockedBy) != 1 {
		t.Fatalf("dependency case = %#v", report.Cases[3])
	}
}

func TestGraphBindingMutationIsRejected(t *testing.T) {
	contract := contractFixture(t)
	graph := graphFixture(contract)
	graph.Nodes[0].ID = "gooo://live-governance-snapshot/input/wrong"
	report := Evaluate(fixtureSnapshot(contract, "normal"), contract, graph)
	if err := ValidateReport(report, contract, graphFixture(contract), "fixture-head"); err == nil {
		t.Fatal("wrong graph entity was accepted")
	}
}

func TestDisabledRulesetAuthorityIsRefuted(t *testing.T) {
	contract := contractFixture(t)
	graph := graphFixture(contract)
	report := Evaluate(fixtureSnapshot(contract, "disabled"), contract, graph)
	if report.Decision != DecisionRefuted || report.Resolution != ResolutionExact {
		t.Fatalf("disabled ruleset report = %#v", report)
	}
	if report.Cases[5].Reason != "DISABLED_RULESET_AUTHORITY" {
		t.Fatalf("disabled case = %#v", report.Cases[5])
	}
}
