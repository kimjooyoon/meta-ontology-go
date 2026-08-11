package freshness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestContractHypothesisCases(t *testing.T) {
	contract := loadFixture(t)
	report, err := Run(contract)
	if err != nil {
		t.Fatal(err)
	}
	if report.HasFailures() {
		t.Fatalf("falsifiable cases failed: %#v", report.Results)
	}
	if report.Passed() {
		t.Fatal("deferred gooo-hosted case was incorrectly reported as pass")
	}
	if !report.HasDeferred() {
		t.Fatal("future self-hosting case was not reported as deferred")
	}
	baseline := findCase(report, "baseline")
	if baseline.Measurement.ActiveRecords != 3 || baseline.Measurement.DependencyEdges != 2 || baseline.Measurement.ProvenanceEdges != 2 || baseline.Measurement.NonFresh != 0 {
		t.Fatalf("baseline measurement=%#v", baseline.Measurement)
	}
	if result := findCase(report, "source-change"); result.Measurement.NonFresh != 2 {
		t.Fatalf("source mutation measurement=%d, want 2", result.Measurement.NonFresh)
	}
	if result := findCase(report, "missing-evidence"); result.Measurement.NonFresh != 1 {
		t.Fatalf("missing evidence measurement=%d, want 1", result.Measurement.NonFresh)
	}
}

func TestContractCanonicalAndReportOutputAreDeterministic(t *testing.T) {
	first := loadFixture(t)
	second := loadFixture(t)
	firstJSON, err := CanonicalJSON(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := CanonicalJSON(second)
	if err != nil || !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("canonical contract changed: err=%v", err)
	}
	var reordered Contract
	data, _ := json.Marshal(first)
	if err := json.Unmarshal(data, &reordered); err != nil {
		t.Fatal(err)
	}
	reverseRecords(reordered.Records)
	reverseCases(reordered.Cases)
	for i := range reordered.Records {
		reverseStrings(reordered.Records[i].InputIDs)
		reverseStrings(reordered.Records[i].Provenance.UsedIDs)
	}
	reorderedJSON, err := CanonicalJSON(reordered)
	if err != nil || !bytes.Equal(firstJSON, reorderedJSON) {
		t.Fatalf("canonical order normalization changed output: err=%v", err)
	}
	firstReport, err := Run(first)
	if err != nil {
		t.Fatal(err)
	}
	secondReport, err := Run(second)
	if err != nil {
		t.Fatal(err)
	}
	firstOutput, _ := json.Marshal(firstReport)
	secondOutput, _ := json.Marshal(secondReport)
	if !bytes.Equal(firstOutput, secondOutput) {
		t.Fatal("experiment output changed between identical runs")
	}
}

func TestContractWrongExpectationFails(t *testing.T) {
	contract := loadFixture(t)
	for i := range contract.Cases {
		if contract.Cases[i].ID != "source-change" {
			continue
		}
		contract.Cases[i].Expected[0].Status = StatusFresh
	}
	report, err := Run(contract)
	if err != nil {
		t.Fatal(err)
	}
	if !report.HasFailures() {
		t.Fatal("counterexample with a false fresh expectation passed")
	}
}

func TestContractRejectsUnknownMutationReference(t *testing.T) {
	contract := loadFixture(t)
	contract.Cases[0].Mutation = Mutation{Kind: "content", RecordID: "billing://missing", Content: "changed"}
	if _, err := Run(contract); err == nil {
		t.Fatal("unknown mutation reference was accepted")
	}
}

func loadFixture(t *testing.T) Contract {
	t.Helper()
	path := filepath.Join("testdata", "freshness-contract.json")
	contract, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	return contract
}

func findCase(report Report, id string) CaseResult {
	for _, result := range report.Results {
		if result.ID == id {
			return result
		}
	}
	return CaseResult{}
}

func reverseRecords(values []Record) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseCases(values []Case) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

func reverseStrings(values []string) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}
