package sourceimpact

import (
	"strings"
	"testing"
)

const fillReferenceSource = `package fillrefs
namespace fillrefs
entity Integer id "fillrefs://integer"
activity Step(Integer) -> Integer id "fillrefs://step" computes "return input + 1"
activity Build(Integer) -> Integer id "fillrefs://build" computes "return __GOOO_BODY_HOLE_value__" assembling {
 source_fill intent "Use a source-declared helper to assemble the value." {
  hole "value"
  candidate "helper" { fill "value" "Step(input)" }
  candidate "identity" { fill "value" "input" }
 }
 case "2" -> "3"
 holdout_case "10" -> "11"
}
activity Main(Integer) -> Integer id "fillrefs://main" computes "return Build(input)"
`

func TestExplicitFillExpressionContributesSourceDependency(t *testing.T) {
	r := impactOf(t, fillReferenceSource, strings.Replace(fillReferenceSource, "return input + 1", "return input + 2", 1))
	if r.Status != "COMPLETE_STATIC" || len(r.Affected) != 3 {
		t.Fatal(r)
	}
	row := r.Affected[0]
	if row.ID != "fillrefs://build" || row.Causes[0].Via.Surface != "fill:helper:value:0" || row.AfterObligations.Holdouts != 1 || row.AfterObligations.FillCandidates != 2 || row.AfterObligations.MaxAttempts != 0 {
		t.Fatal(row)
	}
}

func TestFillExpressionDoesNotInventGlobalEdgeForLocalBinding(t *testing.T) {
	before := strings.Replace(fillReferenceSource, "return __GOOO_BODY_HOLE_value__", "let Step = input; return __GOOO_BODY_HOLE_value__", 1)
	after := strings.Replace(before, "return input + 1", "return input + 2", 1)
	r := impactOf(t, before, after)
	if r.Status != "PARTIAL_STATIC" || len(r.Unknowns) != 2 || len(r.Affected) != 1 || r.Affected[0].ID != "fillrefs://step" {
		t.Fatal(r)
	}
	for _, u := range r.Unknowns {
		if u.ActivityID != "fillrefs://build" || u.Surface != "fill:helper:value:0" || u.Reference != "Step" {
			t.Fatal(u)
		}
	}
}

func TestDerivedCandidateCallSetIsExplicitlyUnresolved(t *testing.T) {
	before := strings.Replace(fillReferenceSource, `candidate "helper" { fill "value" "Step(input)" }
  candidate "identity" { fill "value" "input" }`, `derive grammar "integer-offset-constant/v1" max_expressions "8" max_candidates "16"`, 1)
	r := impactOf(t, before, strings.Replace(before, `max_candidates "16"`, `max_candidates "8"`, 1))
	if r.Status != "PARTIAL_STATIC" || len(r.Unknowns) != 2 || r.Unknowns[0].Surface != "fill_generation" {
		t.Fatal(r)
	}
	if len(r.Affected) != 2 || r.Affected[0].BeforeObligations.GeneratedCandidateLimit != 16 || r.Affected[0].AfterObligations.GeneratedCandidateLimit != 8 {
		t.Fatal(r.Affected)
	}
}

func TestFieldAlternativeUsesEnclosingBindingContext(t *testing.T) {
	before := `package fieldrefs
namespace fieldrefs
entity Integer id "fieldrefs://integer"
entity Box id "fieldrefs://box" fields {
 field value id "fieldrefs://box/value" type integer required one
}
activity Step(Integer) -> Integer id "fieldrefs://step" computes "return input + 1"
activity Build(Integer) -> Box id "fieldrefs://build" computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "Step(input)" intent "Use the helper's value."
 value_case "[2]" -> "{\"value\":3}"
 attempts "2"
}
`
	r := impactOf(t, before, strings.Replace(before, "return input + 1", "return input + 2", 1))
	if r.Status != "COMPLETE_STATIC" || len(r.Affected) != 2 || r.Affected[0].Causes[0].Via.Surface != "choice:value:0" {
		t.Fatal(r)
	}
	before = strings.Replace(before, "return Box{value: input}", "let Step = input; return Box{value: Step}", 1)
	r = impactOf(t, before, strings.Replace(before, "return input + 1", "return input + 2", 1))
	if r.Status != "PARTIAL_STATIC" || len(r.Affected) != 1 || len(r.Unknowns) != 2 || !strings.Contains(r.Unknowns[0].Reason, "enclosing binding") {
		t.Fatal(r)
	}
}
