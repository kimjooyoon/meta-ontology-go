package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"testing"
)

const derivedOrigin26731 = `package derived26731
namespace derived26731
entity Integer id "urn:gooo:type:integer"
// 문법을 기원으로 기록한다.
activity Choose(Integer)->Integer id "urn:derived26731:choose" computes "if __GOOO_BODY_HOLE_condition__ { return \u005f_GOOO_BODY_HOLE_value__ }; return input" assembling {
 source_fill intent "조건과 정수의 기원을 읽는다." {
  hole "condition"
  hole "value"
  derive assignments max_candidates "4" {
   hole "condition" grammar "\u0069nteger-predicate/v1" max_expressions "2"
   hole "value" grammar "integer-offset-constant/v1" max_expressions "3"
  }
 }
 case "23" -> "24"
 case "29" -> "30"
 holdout_case "37" -> "38"
}
`

func TestDerivedOriginPlan26731(t *testing.T) {
	source := []byte(derivedOrigin26731)
	list, err := InspectSourceFillOriginCandidates(context.Background(), "문법26731.gooo", source, "Choose")
	if err != nil || list.Schema != "gooo/source-fill-origin-candidates/v1" || list.Kind != "derived_grammar_assignments" || list.Generation == nil || len(list.Candidates) != 4 {
		t.Fatal(list, err)
	}
	g := list.Generation
	if g.AssignmentSpaceSize != 6 || g.AssignmentsRetained != 4 || g.AssignmentsOmitted != 2 || list.ActivityID != "urn:derived26731:choose" || list.SourceSHA256 != digest(source) {
		t.Fatal("source-owned denominators", list)
	}
	id := list.Candidates[1].CandidateID
	r, err := InspectSourceFillDerivedOrigins(context.Background(), "문법26731.gooo", source, "Choose", id)
	if err != nil || r.Status != "RESOLVED" || r.UnresolvedRows != 0 || len(r.Rows) != 2 || r.PlanSHA256 != list.PlanSHA256 || r.Generation.CandidateSetSHA256 != g.CandidateSetSHA256 {
		t.Fatal("derived assignment source binding", r, err)
	}
	for i, row := range r.Rows {
		raw := string(source[row.GrammarSource.Start:row.GrammarSource.End])
		decoded, err := strconv.Unquote(`"` + raw + `"`)
		if err != nil || decoded != row.Grammar || row.Expression != list.Candidates[1].HoleFills[i].Expression || row.TemplateSource == nil || row.TemplateDecoded == nil {
			t.Fatal("grammar/use mapping", row, raw, err)
		}
		if row.HoleID == "value" && string(source[row.TemplateSource.Start:row.TemplateSource.End]) != `\u005f_GOOO_BODY_HOLE_value__` {
			t.Fatal("escaped template token range", row)
		}
	}
	encoded, err := json.Marshal(r)
	if err != nil || strings.Contains(string(encoded), `"candidate_source"`) || strings.Contains(string(encoded), `"expression_source"`) {
		t.Fatal("derived expression gained a direct declaration range", string(encoded), err)
	}
	raw, err := SourceFillDerivedOriginInputs(r, "origin/derived26731:Inspect")
	var inputs struct {
		Inputs []map[string]SourceFillDerivedInput `json:"inputs"`
	}
	if err != nil || json.Unmarshal(raw, &inputs) != nil || len(inputs.Inputs) != 2 {
		t.Fatal("derived primitive rows", string(raw), err)
	}
	for _, input := range inputs.Inputs {
		row := input["origin/derived26731:Inspect"]
		if row.OriginKind != "derived_grammar" || row.GrammarEnd <= row.GrammarStart || row.ExpressionStart != 0 || row.ExpressionEnd != 0 || !row.HasUse {
			t.Fatal("grammar origin masqueraded as a direct expression", row)
		}
	}
	if _, err := InspectSourceFillDerivedOrigins(context.Background(), "文法26731.gooo", source, "Choose", "not_retained26731"); err == nil || !strings.Contains(err.Error(), "retained source assignment set") {
		t.Fatal("unknown assignment gained provenance", err)
	}
	t.Log("grammar26731:4 retained assignments of6;2 omitted;2 original grammar/hole/use rows for selected lookup;no expression/candidate declaration fabricated;producer/model0")
}

func TestDerivedOriginUnresolved26731(t *testing.T) {
	source := []byte(strings.Replace(derivedOrigin26731, `return \u005f_GOOO_BODY_HOLE_value__`, `return input`, 1))
	list, err := InspectSourceFillOriginCandidates(context.Background(), "无使用26731.gooo", source, "Choose")
	if err != nil {
		t.Fatal(err)
	}
	r, err := InspectSourceFillDerivedOrigins(context.Background(), "无使用26731.gooo", source, "Choose", list.Candidates[0].CandidateID)
	if err != nil || r.Status != "UNRESOLVED" || r.UnresolvedRows != 1 || r.Rows[1].TemplateSource != nil || r.Rows[1].Reason != "TEMPLATE_HOLE_NOT_UNIQUE" || r.Rows[1].GrammarSource.End <= r.Rows[1].GrammarSource.Start {
		t.Fatal("unresolved use lost original grammar", r, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := InspectSourceFillOriginCandidates(ctx, "cancel26731.gooo", source, "Choose"); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled planning progressed", err)
	}
	if _, err := InspectSourceFillDerivedOrigins(nil, "nil26731.gooo", source, "Choose", "candidate26731"); err == nil {
		t.Fatal("missing context progressed")
	}
	t.Log("grammar26731:missing use retains grammar with no use range;unknown assignment/context/cancellation remain errors;model0")
}
