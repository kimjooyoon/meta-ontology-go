package bodycodegen

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

const twoRecordQuestions = `package limits
namespace limits
entity Integer id "limits://integer"
entity Pair id "limits://pair" fields {
 field a id "limits://a" type integer required one
 field b id "limits://b" type integer required one
}
activity Choose(Integer,Integer) -> Pair computes "return Pair{a: input0, b: input1}" assembling {
 choice "a" field_value at "0" alternative "0 - input0" intent "Negate first."
 choice "b" field_value at "1" alternative "0 - input1" intent "Negate second."
 value_case "[0,0]" -> "{\"a\":0,\"b\":0}"
 observe { value_input "[7,0]" value_input "[0,9]" oracle "Expected" max_candidates "4" max_rounds "1" resolve_unique "true" }
 attempts "4"
}
activity Expected(Integer,Integer) -> Pair computes "return Pair{a: 0 - input0, b: 0 - input1}"
`

func TestRecordObservationRoundLimitLeavesAQuestion(t *testing.T) {
	g, _ := NewTypedPathGenerator("")
	r, err := g.GenerateSourceAssembly(context.Background(), "rounds.gooo", []byte(twoRecordQuestions), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	o := r.Report.RecordAssembly.Observation
	if o.Status != "ROUND_BUDGET" || o.ResolvedMask != nil || o.EffectiveCases != 2 || len(o.Rounds) != 2 ||
		len(o.Rounds[1].SurvivingMasks) != 2 || o.Rounds[1].RecommendedIndex == nil || *o.Rounds[1].RecommendedIndex != 1 || o.Rounds[1].Observation != nil {
		t.Fatal(o)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "rounds.gooo", []byte(twoRecordQuestions), r); err != nil {
		t.Fatal(err)
	}
}

func TestRecordObservationUnavailableCandidateCannotResolveUniquely(t *testing.T) {
	source := `package unknown
namespace unknown
entity Integer id "unknown://integer"
entity Box id "unknown://box" fields { field value id "unknown://value" type integer required one }
activity Choose(Integer) -> Box computes "return Box{value: input}" assembling {
 choice "value" field_value at "0" alternative "1 / input" intent "Question the reciprocal."
 value_case "[0]" -> "{\"value\":0}"
 observe { value_input "[7]" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}`
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "unknown.gooo", []byte(source), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	o := result.Report.RecordAssembly.Observation
	if o.Status != "EVALUATION_UNAVAILABLE" || o.ResolvedMask != nil || len(o.Rounds[0].UnavailableMasks) != 1 ||
		len(o.Rounds[0].Unavailable) != 1 || o.Rounds[0].Unavailable[0].Error == "" {
		t.Fatal(o)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "unknown.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordObservationCaseLimitPreservesSeparateHoldout(t *testing.T) {
	source := `package casebound
namespace casebound
entity Integer id "casebound://integer"
entity Box id "casebound://box" fields { field value id "casebound://value" type integer required one }
activity Choose(Integer,Integer) -> Box computes "return Box{value: input0}" assembling {
 choice "value" field_value at "0" alternative "0 - input0" intent "Negate first."
 CASES
 holdout_value_case "[61,0]" -> "{\"value\":-61}"
 observe { value_input "[7,0]" oracle "Expected" max_candidates "2" max_rounds "1" }
 attempts "2"
}
activity Expected(Integer,Integer) -> Box computes "return Box{value: 0 - input0}"
`
	var cases strings.Builder
	for i := range 127 {
		fmt.Fprintf(&cases, "value_case \"[0,%d]\" -> \"{\\\"value\\\":0}\"\n", i)
	}
	source = strings.Replace(source, "CASES", cases.String(), 1)
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "casebound.gooo", []byte(source), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.Observation.Status != "CASE_BUDGET" || r.Total != 127 || r.Observation.OracleEvaluations != 127 || r.Observation.Rounds[0].Observation != nil || r.Holdout.Total != 1 {
		t.Fatal(r)
	}
}
