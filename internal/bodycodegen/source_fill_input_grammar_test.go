package bodycodegen

import "testing"

func TestOriginalGrammarInputOrigins4741(t *testing.T) {
	const source = `package grade4741
namespace grade4741
entity Reading id "urn:grade4741:reading" fields {
 field score id "urn:grade4741:score" type integer required one
 field zone id "urn:grade4741:zone" type string required one
}
entity Text id "urn:gooo:type:string"
activity Classify(Reading) -> Text computes "if __GOOO_BODY_HOLE_gate__ { return __GOOO_BODY_HOLE_label__ }; return \"review4741\"" assembling {
 source_fill intent "측정 조건과 설명 문구의 기원을 따로 기록한다" {
  hole "gate"
  hole "label"
  derive assignments max_candidates "4" {
   hole "gate" grammar "record-field-predicate/v2" max_expressions "2"
   hole "label" grammar "record-string-literal/v1" max_expressions "2"
  }
 }
 value_case "[{\"score\":83,\"zone\":\"indoor\"}]" -> "\"accepted4741\""
}
`
	decl, err := InspectSourceFillDeclarationOrigins("측정4741.gooo", []byte(source), "Classify", "derived4741")
	if err != nil || len(decl.Holes) != 2 || decl.CandidateSource != nil {
		t.Fatal(err, decl)
	}
	want := map[string]string{"gate": "record-field-predicate/v2", "label": "record-string-literal/v1"}
	for _, hole := range decl.Holes {
		if hole.Grammar != want[hole.ID] || hole.ExpressionSource != nil || hole.GrammarSource == nil ||
			source[hole.GrammarSource.Start:hole.GrammarSource.End] != want[hole.ID] {
			t.Fatal("grammar source", hole)
		}
	}
	t.Log("grammar4741: two distinct original per-hole grammar ranges; generated expressions have no authored expression or candidate position; static syntax only; model0")
}
