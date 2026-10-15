package sourceimpact

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func impactOf(t *testing.T, before, after string) Report {
	t.Helper()
	r, err := Compare(context.Background(), "before.gooo", []byte(before), "after.gooo", []byte(after))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDisplayRenamePreservesStableIdentity(t *testing.T) {
	after := strings.ReplaceAll(sourceFixture, "Reference", "ExpectedFee")
	r := impactOf(t, sourceFixture, after)
	if len(r.Changes) != 2 || len(r.Affected) != 3 {
		t.Fatal(r)
	}
	for _, change := range r.Changes {
		for _, facet := range change.Facets {
			if facet == "added" || facet == "removed" {
				t.Fatal("rename replaced stable identity", change)
			}
		}
	}
	if r.Changes[1].ID != "changed://reference" || !reflect.DeepEqual(r.Changes[1].Facets, []string{"presentation"}) {
		t.Fatal(r.Changes)
	}
}

func TestFieldPresenceChangeReachesConsumersAndCallers(t *testing.T) {
	before := `package record
namespace record
entity Integer id "record://integer"
entity Record id "record://record" fields {
 field amount id "record://record/amount" type integer required one
}
activity Value(Record) -> Integer id "record://value" computes "return input.amount"
activity Main(Record) -> Integer id "record://main" computes "return Value(input)"
activity Unrelated(Integer) -> Integer id "record://unrelated" computes "return input"
`
	r := impactOf(t, before, strings.Replace(before, "required one", "optional one", 1))
	if len(r.Changes) != 1 || !reflect.DeepEqual(r.Changes[0].Facets, []string{"fields"}) || len(r.Affected) != 3 {
		t.Fatal(r)
	}
	for _, row := range r.Affected {
		if row.ID == "record://unrelated" || row.ID == "record://integer" {
			t.Fatal("shared type leaked impact", row)
		}
	}
}

func TestRuntimeBindingAndFeedbackKeepDirectionAndKind(t *testing.T) {
	before := `package flow
namespace flow
entity Integer id "flow://integer"
activity Start(Integer) -> Integer id "flow://start" computes "return input + 1"
activity Finish(Integer) -> Integer id "flow://finish" computes "return input - 1"
activity Unrelated(Integer) -> Integer id "flow://unrelated" computes "return input"
bind Start.result -> Finish.input
feedback Finish.result -> Start.input
`
	r := impactOf(t, before, strings.Replace(before, "return input - 1", "return input - 2", 1))
	if len(r.Affected) != 2 || len(r.Changes) != 1 {
		t.Fatal(r)
	}
	for _, row := range r.Affected {
		if row.ID == "flow://start" && row.Causes[1].Via.Kind != "runtime_feedback" {
			t.Fatal(row)
		}
	}
	// The same endpoints with a different delivery phase are a source change.
	withoutFeedback := strings.Replace(before, "feedback Finish.result -> Start.input\n", "", 1)
	ordinary := strings.Replace(withoutFeedback, "bind Start.result -> Finish.input", "feedback Start.result -> Finish.input", 1)
	r = impactOf(t, withoutFeedback, ordinary)
	if len(r.Changes) != 1 || r.Changes[0].ID != "flow://finish" || !reflect.DeepEqual(r.Changes[0].Facets, []string{"dependencies"}) {
		t.Fatal(r)
	}
}

func TestPolicyChangesStayExplicitWhenConsumersAreUnknown(t *testing.T) {
	before := `package policy
namespace policy
policy Choice id "policy://choice" {
 state "START"
 state "PASS"
 state "UNKNOWN"
 transition "START" -> "PASS"
}
`
	r := impactOf(t, before, strings.Replace(before, `transition "START" -> "PASS"`, `transition "START" -> "UNKNOWN"`, 1))
	if r.Status != "PARTIAL_STATIC" || len(r.ScopeChanges) != 1 || r.ScopeChanges[0].Field != "policy:policy://choice" || len(r.Unknowns) != 2 {
		t.Fatal(r)
	}
}

func TestInputCallCannotBecomeGlobalActivityDependency(t *testing.T) {
	before := `package scope
namespace scope
entity Integer id "scope://integer"
activity input(Integer) -> Integer id "scope://shadow" computes "return input + 1"
activity Main(Integer) -> Integer id "scope://main" computes "return input(input)"
`
	r := impactOf(t, before, strings.Replace(before, "return input + 1", "return input + 2", 1))
	if r.Status != "PARTIAL_STATIC" || len(r.Affected) != 1 || r.Affected[0].ID != "scope://shadow" || len(r.Unknowns) != 2 {
		t.Fatal(r)
	}
}

func TestImportedScopeRemainsUnresolvedAndChangesAreVisible(t *testing.T) {
	before := strings.Replace(sourceFixture, "namespace changed\n", "namespace changed\nimport upstream \"example/first\"\n", 1)
	after := strings.Replace(before, "example/first", "example/second", 1)
	r := impactOf(t, before, after)
	if r.Status != "PARTIAL_STATIC" || len(r.ScopeChanges) != 1 || r.ScopeChanges[0].Field != "imports" || len(r.Unknowns) != 2 {
		t.Fatal(r)
	}
}

func TestImpactRejectsMissingContextInvalidSourceAndSourceBudget(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Compare(ctx, "a", []byte(sourceFixture), "b", []byte(sourceFixture)); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, err := Compare(nil, "a", nil, "b", nil); err == nil {
		t.Fatal("nil context accepted")
	}
	for _, source := range []string{"broken", strings.Repeat(" ", maxSourceBytes+1)} {
		if _, err := Compare(context.Background(), "a", []byte(source), "b", []byte(sourceFixture)); err == nil {
			t.Fatal("invalid source accepted")
		}
	}
}

func TestPositionalInputOrderSurvivesUnchangedDependencySet(t *testing.T) {
	before := `package ordered
namespace ordered
entity Integer id "ordered://integer"
entity Budget id "ordered://budget" fields {
 field amount id "ordered://budget/amount" type integer required one
}

entity Charge id "ordered://charge" fields {
 field amount id "ordered://charge/amount" type integer required one
}
activity Difference(Budget, Charge) -> Integer id "ordered://difference" computes "return input0.amount - input1.amount"
`
	r := impactOf(t, before, strings.Replace(before, "Difference(Budget, Charge)", "Difference(Charge, Budget)", 1))
	if len(r.Changes) != 1 || !reflect.DeepEqual(r.Changes[0].Facets, []string{"input_sequence"}) || len(r.Affected) != 1 {
		t.Fatal(r)
	}
}

func TestSignatureWithoutBodyRetainsUninspectedImplementation(t *testing.T) {
	source := `package signature
namespace signature
entity Integer id "signature://integer"
activity Main(Integer) -> Integer id "signature://main"
`
	r := impactOf(t, source, source)
	if r.Status != "PARTIAL_STATIC" || len(r.Unknowns) != 2 || len(r.Changes) != 0 {
		t.Fatal(r)
	}
	for _, unknown := range r.Unknowns {
		if unknown.ActivityID != "signature://main" || unknown.Surface != "computes" || !strings.Contains(unknown.Reason, "source body unavailable") {
			t.Fatal(unknown)
		}
	}
}
