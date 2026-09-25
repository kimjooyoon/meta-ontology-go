package semanticdelta

import (
	"errors"
	"reflect"
	"testing"
)

func TestDetectReportsEveryOutOfScopeEndpointDeterministically(t *testing.T) {
	delta := Delta{
		AddedNodes: []Node{{ID: "billing://entity/other", Kind: "Entity"}},
		AddedFacts: []Fact{{
			Subject: "billing://activity/pay-order", Predicate: "gooo:invokes", Object: "fraud://activity/check",
		}},
	}
	scope := Scope{IDs: []string{"billing://activity/pay-order"}, Predicates: []string{"gooo:invokes"}}
	report, err := Detect(delta, scope)
	if err != nil {
		t.Fatal(err)
	}
	if report.Passes() || len(report.Violations) != 2 {
		t.Fatalf("report = %#v, want two violations", report)
	}
	if report.Violations[0].Change != ChangeNode || report.Violations[1].Endpoint != "object" {
		t.Fatalf("violations are not sorted or classified: %#v", report.Violations)
	}
	repeated, err := Detect(delta, scope)
	if err != nil || !reflect.DeepEqual(report, repeated) {
		t.Fatalf("detection was not deterministic: %#v != %#v (%v)", report, repeated, err)
	}
}

func TestDetectAllowsExactIDsPrefixesAndUnrestrictedPredicates(t *testing.T) {
	delta := Delta{
		AddedNodes: []Node{{ID: "billing://entity/order", Kind: "Entity"}},
		RemovedFacts: []Fact{{
			Subject: "billing://activity/pay-order", Predicate: "prov:used", Object: "billing://entity/order",
		}},
	}
	scope := Scope{Prefixes: []string{"billing://"}}
	report, err := Detect(delta, scope)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Passes() {
		t.Fatalf("prefix scope rejected an in-scope delta: %#v", report)
	}
	if !scope.AllowsPredicate("any:predicate") {
		t.Fatal("empty predicate scope should allow every predicate")
	}
}

func TestDetectRejectsMalformedSemanticItems(t *testing.T) {
	_, err := Detect(Delta{AddedFacts: []Fact{{Subject: "has whitespace", Predicate: "uses", Object: "ok"}}}, Scope{})
	if err == nil {
		t.Fatal("malformed fact was accepted")
	}
	if errors.Is(err, ErrScopeViolation) {
		t.Fatalf("malformed input was reported as a scope violation: %v", err)
	}
}

func TestScopeErrorUnwrapsAndRetainsReport(t *testing.T) {
	report := Report{Allowed: false, Violations: []Violation{{Reason: "outside"}}}
	err := &ScopeError{Report: report}
	if !errors.Is(err, ErrScopeViolation) || !reflect.DeepEqual(err.Report, report) {
		t.Fatalf("scope error = %v %#v", err, err.Report)
	}
}
