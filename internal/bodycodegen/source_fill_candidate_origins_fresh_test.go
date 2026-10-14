package bodycodegen

import (
	"encoding/json"
	"strings"
	"testing"
)

const sourceCandidate24573 = `package candidate24573
namespace candidate24573
entity Integer id "urn:candidate24573:integer"
// 기원은 선택과 별개다.
activity Choose(Integer) -> Integer computes "label := \"__GOOO_BODY_HOLE_value__\"; // __GOOO_BODY_HOLE_value__\nif input < 0 { return __GOOO_BODY_HOLE_bias__ }; return \u005f_GOOO_BODY_HOLE_value__" assembling {
 source_fill intent "Keep a declared source assignment." {
  hole "value"
  hole "bias"
  candidate "wrong_type" { fill "value" "true" fill "bias" "input - 7" }
  candidate "next" { fill "value" "input + 7" fill "bias" "input - 7" }
 }
 case "5" -> "12"
}
`

func TestSourceCandidateOrigins24573(t *testing.T) {
	source := []byte(sourceCandidate24573)
	r, err := InspectSourceFillCandidateOrigins("후보24573.gooo", source, "Choose", "wrong_type")
	if err != nil || r.Status != "RESOLVED" || r.UnresolvedRows != 0 || len(r.Rows) != 2 || r.ActivityID == "" || r.SourceSHA256 != digest(source) {
		t.Fatal(r, err)
	}
	if string(source[r.CandidateSource.Start:r.CandidateSource.End]) != "wrong_type" {
		t.Fatal("candidate declaration range", r.CandidateSource)
	}
	for _, row := range r.Rows {
		if row.Status != "MAPPED" || row.TemplateSource == nil || row.TemplateDecoded == nil || row.Reason != "" {
			t.Fatal("unselected candidate lost its source use", row)
		}
		if string(source[row.ExpressionSource.Start:row.ExpressionSource.End]) != row.Expression || !strings.Contains(string(source[row.HoleDeclaration.Start:row.HoleDeclaration.End]), row.HoleID) {
			t.Fatal("declaration expression/hole location", row)
		}
		use := string(source[row.TemplateSource.Start:row.TemplateSource.End])
		if row.HoleID == "value" && use != `\u005f_GOOO_BODY_HOLE_value__` || row.HoleID == "bias" && use != bodyFillHoleToken("bias") {
			t.Fatal("raw literal offsets or comment/string token exclusion", use)
		}
	}
	if r.Rows[0].Expression != "true" {
		t.Fatal("declaration inspection invented a type-checked replacement")
	}
	raw, err := SourceFillCandidateOriginInputs(r, "origin/candidate24573:Inspect")
	var input struct {
		Inputs []map[string]SourceFillCandidateInput `json:"inputs"`
	}
	if err != nil || json.Unmarshal(raw, &input) != nil || len(input.Inputs) != 2 || !input.Inputs[0]["origin/candidate24573:Inspect"].HasUse {
		t.Fatal("primitive rows", string(raw), err)
	}
	for _, target := range []string{"Inspect", "origin:Inspect:Other", "origin :Inspect"} {
		if _, err := SourceFillCandidateOriginInputs(r, target); err == nil {
			t.Fatal("malformed target accepted", target)
		}
	}
	if _, err := InspectSourceFillCandidateOrigins("missing24573.gooo", source, "Choose", "absent"); err == nil {
		t.Fatal("absent candidate gained declared origins")
	}
	t.Log("source24573:2 explicit hole mappings, escaped/Unicode raw byte locations, invalid-for-Integer expression retained; no candidate/model execution")
}

func TestSourceCandidateUnresolved24573(t *testing.T) {
	for _, replacement := range []string{`return input`, `return __GOOO_BODY_HOLE_value__ + __GOOO_BODY_HOLE_value__`} {
		source := []byte(strings.Replace(sourceCandidate24573, `return \u005f_GOOO_BODY_HOLE_value__`, replacement, 1))
		r, err := InspectSourceFillCandidateOrigins("不确定24573.gooo", source, "Choose", "next")
		if err != nil || r.Status != "UNRESOLVED" || r.UnresolvedRows != 1 || len(r.Rows) != 2 {
			t.Fatal(r, err)
		}
		row := r.Rows[0]
		if row.TemplateSource != nil || row.TemplateDecoded != nil || row.Reason != "TEMPLATE_HOLE_NOT_UNIQUE" || r.Rows[1].Status != "MAPPED" {
			t.Fatal("absent/ambiguous use acquired a location", row)
		}
		raw, err := SourceFillCandidateOriginInputs(r, "origin/candidate24573:Inspect")
		var input struct {
			Inputs []map[string]SourceFillCandidateInput `json:"inputs"`
		}
		if err != nil || json.Unmarshal(raw, &input) != nil || input.Inputs[0]["origin/candidate24573:Inspect"].HasUse {
			t.Fatal("unresolved primitive row gained a use", err)
		}
	}
	oversized := SourceFillCandidateOrigins{Schema: "gooo/source-fill-candidate-origins/v1", Rows: []SourceFillCandidateOrigin{{Expression: strings.Repeat("x", 33<<10)}}}
	if _, err := SourceFillCandidateOriginInputs(oversized, "origin/candidate24573:Inspect"); err == nil {
		t.Fatal("oversized export silently truncated")
	}
	t.Log("source24573:absent and repeated uses remain explicit unresolved observations; no candidate/model execution")
}
