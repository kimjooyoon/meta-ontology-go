package bodycodegen

import "testing"

func TestOriginalFillDeclarations4571(t *testing.T) {
	const source = `package authored4571
namespace authored4571
import 値 "origin4571/value"
activity Build(値.Packet) -> 値.Packet computes "return __GOOO_BODY_HOLE_value__" assembling {
 source_fill intent "원래 생성자와 문자열을 남긴다" {
  hole "value"
  candidate "packet4571" { fill "value" "値.Packet{note: \"발송\", count: input.count + 37}" }
  candidate "unchanged4571" { fill "value" "input" }
 }
 value_case "[{\"note\":\"a\",\"count\":4}]" -> "{\"note\":\"발송\",\"count\":41}"
}
entity Text id "urn:gooo:type:string"
activity Route(値.Packet) -> Text computes "if __GOOO_BODY_HOLE_gate__ { return __GOOO_BODY_HOLE_packet__ }; return \"hold4571\"" assembling {
 source_fill intent "각 문법의 출처를 분리한다" {
  hole "gate"
  hole "packet"
  derive assignments max_candidates "4" {
   hole "gate" grammar "record-field-predicate/v2" max_expressions "2"
   hole "packet" grammar "record-string-literal/v1" max_expressions "2"
  }
 }
 value_case "[{\"note\":\"a\",\"count\":4}]" -> "\"send4571\""
}
`
	decl, err := InspectSourceFillDeclarationOrigins("기원4571.gooo", []byte(source), "Build", "packet4571")
	if err != nil || len(decl.Holes) != 1 || decl.CandidateSource == nil {
		t.Fatal("explicit declaration", err, decl)
	}
	hole := decl.Holes[0]
	if hole.Expression != `値.Packet{note: "발송", count: input.count + 37}` || hole.ExpressionSource == nil ||
		source[hole.ExpressionSource.Start:hole.ExpressionSource.End] != `値.Packet{note: \"발송\", count: input.count + 37}` {
		t.Fatal("authored constructor range", hole)
	}
	derived, err := InspectSourceFillDeclarationOrigins("기원4571.gooo", []byte(source), "Route", "derived4571")
	if err != nil || derived.CandidateSource != nil || len(derived.Holes) != 2 {
		t.Fatal("derived declaration", err, derived)
	}
	for _, item := range derived.Holes {
		if item.ExpressionSource != nil || item.GrammarSource == nil || source[item.GrammarSource.Start:item.GrammarSource.End] != item.Grammar {
			t.Fatal("grammar location", item)
		}
	}
	if _, err := InspectSourceFillDeclarationOrigins("기원4571.gooo", []byte(source), "Build", "absent4571"); err == nil {
		t.Fatal("absent candidate acquired a source")
	}
	t.Log("declarations4571: static syntax only; escaped constructor source retained; two per-hole grammar locations; absent candidate unresolved; no candidate generation or native execution")
}
