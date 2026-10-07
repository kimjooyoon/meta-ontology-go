package bodycodegen

import (
	"context"
	"strings"
	"testing"
)

func TestRecordPolicyUsesBestWholeCaseCandidate(t *testing.T) {
	source := string(recordUpdatesFixture(t))
	source = source[:strings.Index(source, "activity Select")] + `activity Select(Candidate, Boolean) -> Candidate computes ` + "`" + `return Candidate{title: "a", state: "a", reason: "a"}` + "`" + ` assembling {
    choice "title" field_value at "0" alternative "\"b\"" intent "title"
    choice "state" field_value at "1" alternative "\"b\"" intent "state"
    choice "reason" field_value at "2" alternative "\"b\"" intent "reason"
    value_case "[{\"title\":\"x\",\"state\":\"x\",\"reason\":\"x\"},true]" -> "{\"title\":\"a\",\"state\":\"a\",\"reason\":\"a\"}"
    value_case "[{\"title\":\"y\",\"state\":\"y\",\"reason\":\"y\"},true]" -> "{\"title\":\"b\",\"state\":\"b\",\"reason\":\"b\"}"
    value_case "[{\"title\":\"z\",\"state\":\"z\",\"reason\":\"z\"},true]" -> "{\"title\":\"b\",\"state\":\"b\",\"reason\":\"b\"}"
    attempts "8"
}`
	g, _ := NewTypedPathGenerator("")
	result, err := g.GenerateRecordAssemblyWithPolicy(context.Background(), "tradeoff.gooo", []byte(source), "Select", assemblyPolicyFixture(t))
	if err != nil {
		t.Fatal(err)
	}
	r := result.Report.RecordAssembly
	if len(r.Attempts) != 2 || r.SelectedMask != 0 || r.Passed != 1 || r.FieldsPassed != 3 ||
		r.Attempts[1].Passed != 0 || r.Attempts[1].FieldsPassed != 4 ||
		r.Control.Decisions[1].Operation != "USE_OBSERVED_CANDIDATE" {
		t.Fatal("whole-case policy selected a field-only improvement", r)
	}
	if _, err := RealizeSourceAssembly(context.Background(), "tradeoff.gooo", []byte(source), result); err != nil {
		t.Fatal(err)
	}
}
