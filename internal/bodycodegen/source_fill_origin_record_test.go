package bodycodegen

import (
	"context"
	"testing"
)

func TestRecordFillPerHoleOriginAndAbsentHistory3751(t *testing.T) {
	source := []byte(`package record3751
namespace record3751
entity Parcel id "urn:record3751:parcel" fields {
 field zone id "urn:record3751:zone" type string required one
 field packed id "urn:record3751:packed" type boolean required one
}
entity Label id "urn:record3751:label" fields {
 field route id "urn:record3751:route" type string required one
}
activity Dispatch(Parcel) -> Label id "urn:record3751:dispatch" computes "if __GOOO_BODY_HOLE_gate__ { return Label{route: __GOOO_BODY_HOLE_label__} }; return Label{route: \"hold3751\"}" assembling {
 source_fill intent "택배 조건과 라벨을 만든 문법을 따로 기록한다" {
  hole "label"
  hole "gate"
  derive assignments max_candidates "4" {
   hole "label" grammar "record-string-literal/v1" max_expressions "2"
   hole "gate" grammar "record-field-predicate-composition/v1" max_expressions "2"
  }
 }
 value_case "[{\"zone\":\"east3751\",\"packed\":true}]" -> "{\"route\":\"ship3751\"}"
 value_case "[{\"zone\":\"east3751\",\"packed\":false}]" -> "{\"route\":\"hold3751\"}"
}
`)
	prior := generateOriginFill3371(t, "record3751.gooo", source, "Dispatch")
	origin := prior.Report.BodyFill.SourceOrigin
	if origin == nil || origin.CandidateSource != nil {
		t.Fatal("record derivation origin missing")
	}
	grammars := map[string]string{"label": "record-string-literal/v1", "gate": "record-field-predicate-composition/v1"}
	derived := 0
	for _, fragment := range origin.Fragments {
		if fragment.HoleID == "" {
			continue
		}
		derived++
		if fragment.Kind != "derived_grammar" || fragment.Grammar != grammars[fragment.HoleID] ||
			fragment.ExpressionSource != nil || fragment.GrammarSource == nil ||
			string(source[fragment.GrammarSource.Start:fragment.GrammarSource.End]) != fragment.Grammar {
			t.Fatal("per-hole grammar origin was mixed", fragment)
		}
	}
	if derived != 2 {
		t.Fatal("per-hole origins", derived)
	}
	saveFillOrigin3371(t, "record3751", source, prior)
	// This is a newly generated record with an intentionally absent optional
	// origin, not a replay of a historical source or saved fixture.
	prior.Report.BodyFill.SourceOrigin = nil
	if _, err := RealizeSourceAssembly(context.Background(), "record3751.gooo", source, prior); err != nil {
		t.Fatal("explicit origin absence could not be read", err)
	}
	if prior.Report.BodyFill.SourceOrigin != nil {
		t.Fatal("comparison fabricated past provenance")
	}
	t.Log("record3751: two distinct per-hole grammars and exact source positions; absent optional origin stays absent; typed generation only; model0")
}
