package semanticdelta

import (
	"embed"
	"reflect"
	"testing"
)

//go:embed testdata/negative-*.json testdata/negative-*.txt testdata/expected-negative-*.json
var negativeFixtureFiles embed.FS

const (
	semanticDeltaHypothesis = "H1: canonical detection rejects every changed endpoint outside the declared scope"
	semanticDeltaPassRule   = "pass when the normalized report is allowed with zero violations"
	semanticDeltaFailRule   = "fail when a node, subject, predicate, or object violation is missing"
	semanticDeltaDeferred   = "defer gooo-hosted execution and promotion until that host exists"
)

func TestNegativeEndpointFixtureRejectsEveryOutOfScopeEndpoint(t *testing.T) {
	t.Logf("hypothesis=%s pass=%s fail=%s deferred=%s", semanticDeltaHypothesis, semanticDeltaPassRule, semanticDeltaFailRule, semanticDeltaDeferred)
	request := loadNegativeFixture(t, "negative-endpoints.json")
	report, err := Detect(request.Delta, request.Allowed)
	if err != nil {
		t.Fatal(err)
	}
	want := Report{Allowed: false, Violations: []Violation{
		{
			Operation: OperationAdd, Change: ChangeNode, ID: "fraud://entity/charge", Kind: "Entity",
			Endpoint: "node", Reason: "node identity is outside allowed scope",
		},
		{
			Operation: OperationAdd, Change: ChangeFact, Subject: "billing://activity/pay-order",
			Predicate: "gooo:invokes", Object: "fraud://entity/charge", Endpoint: "object",
			Reason: "fact object is outside allowed scope",
		},
		{
			Operation: OperationAdd, Change: ChangeFact, Subject: "billing://activity/pay-order",
			Predicate: "gooo:invokes", Object: "fraud://entity/charge", Endpoint: "predicate",
			Reason: "fact predicate is outside allowed scope",
		},
	}}
	if !reflect.DeepEqual(report, want) {
		t.Fatalf("negative endpoint report = %#v, want %#v", report, want)
	}
	assertNegativeReportFixture(t, "negative-endpoints.json", report, "expected-negative-endpoints.json")
}
func TestNegativeRemovalFixtureTreatsRemovalAsExplicitDelta(t *testing.T) {
	request := loadNegativeFixture(t, "negative-removal.txt")
	report, err := Detect(request.Delta, request.Allowed)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passes() || len(report.Violations) != 2 {
		t.Fatalf("removal fixture report = %#v, want two violations", report)
	}
	if report.Violations[0].Operation != OperationRemove || report.Violations[1].Endpoint != "object" {
		t.Fatalf("removal violations lost operation or endpoint: %#v", report.Violations)
	}
	assertNegativeReportFixture(t, "negative-removal.txt", report, "expected-negative-removal.json")
}
func TestNegativeMalformedFixtureFailsBeforeDetection(t *testing.T) {
	data, err := negativeFixtureFiles.ReadFile("testdata/negative-malformed.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeJSON(data); err == nil {
		t.Fatal("malformed fixture was accepted")
	}
}
