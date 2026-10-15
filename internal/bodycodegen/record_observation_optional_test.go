package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestRecordObservationDistinguishesPresentZeroFromAbsence(t *testing.T) {
	source := `package presence
namespace presence
entity Boolean id "presence://boolean"
entity Box id "presence://box" fields { field value id "presence://value" type integer optional one }
activity Choose(Boolean) -> Box computes "let value = 0; let result = Box{}; if input { result.value = &value }; return result" assembling {
 choice "value" field_update at "0" alternative "nil" intent "Preserve an explicit zero."
 value_case "[false]" -> "{}"
 observe { value_input "[true]" oracle "Expected" max_candidates "2" max_rounds "1" resolve_unique "true" }
 attempts "2"
}
activity Expected(Boolean) -> Box computes "let value = 0; let result = Box{}; if input { result.value = &value }; return result"
`
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateSourceAssembly(context.Background(), "presence.gooo", []byte(source), "Choose")
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if r.SelectedMask != 0 || r.Total != 2 || r.Passed != 2 || r.Observation.ResolvedMask == nil ||
		r.Observation.Rounds[0].Observation.Expected != `{"value":0}` || r.Cases[1].Fields[0].ActualPresent == nil || !*r.Cases[1].Fields[0].ActualPresent {
		t.Fatal(r)
	}
	if _, err = RealizeSourceAssembly(context.Background(), "presence.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}

func TestRecordObservationRejectsUnresolvedReferenceClosure(t *testing.T) {
	source := strings.Replace(sourceRecordObservationFixture,
		`let result = input0; if input1 { result.amount = 0 - input0.amount; result.approved = true }; return result`,
		`return Decide(input0,input1)`, 1)
	g, _ := NewTypedPathGenerator("")
	if _, err := g.GenerateSourceAssembly(context.Background(), "reference.gooo", []byte(source), "Decide"); err == nil {
		t.Fatal("assembling callee supplied the oracle answer")
	}
}
