package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/jointdecision"
)

func flowSource(t *testing.T, body string) []byte {
	t.Helper()
	source := string(recordUpdatesFixture(t))
	start := strings.Index(source, "computes `") + len("computes `")
	end := strings.Index(source, "` assembling {")
	return []byte(source[:start] + body + source[end:])
}

func exportedFlow(t *testing.T, source []byte) RecordAssemblyContextExport {
	t.Helper()
	r, err := ExportRecordAssemblyContextWithFlow(context.Background(), "flow.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
	if err != nil || r.ValueFlow == nil || r.ValueFlow.Status != "RESOLVED" {
		t.Fatalf("flow export: %+v, %v", r.ValueFlow, err)
	}
	if r.ModelPredictions != 0 || r.CandidateTests != 0 || r.ExpandedPlan != nil {
		t.Fatal("source analysis executed or exported cases")
	}
	return r
}

func flowContains(f *RecordValueFlow, root uint16, kind, operator string) bool {
	if root == 0 {
		return false
	}
	node := f.Nodes[root-1]
	return (node.Kind == kind && (operator == "" || operator == node.Operator)) ||
		flowContains(f, node.Parents[0], kind, operator) || flowContains(f, node.Parents[1], kind, operator)
}

func TestRecordFlowDistinguishesCurrentFieldsAndRecordScalarSnapshots(t *testing.T) {
	body := "let copy = input0\nlet saved = copy\nlet old = copy.state\ncopy.title = \"draft\"\ncopy.state = \"wait\"\ncopy.reason = old\nreturn Candidate{title: copy.title, state: copy.state, reason: copy.reason + saved.reason + old}"
	source := flowSource(t, body)
	current := exportedFlow(t, source)
	savedSource := []byte(strings.Replace(string(source), `copy.title + \":\" + copy.state`, `saved.title + \":\" + saved.state`, 1))
	saved := exportedFlow(t, savedSource)
	a, b := current.ValueFlow.Choices[2], saved.ValueFlow.Choices[2]
	if flowContains(current.ValueFlow, a.First, "choice", "state") || !flowContains(current.ValueFlow, a.Second, "choice", "state") {
		t.Fatal("scalar snapshot followed a later state write or current read lost it")
	}
	if flowContains(saved.ValueFlow, b.Second, "choice", "state") || !flowContains(saved.ValueFlow, b.Second, "input", "") {
		t.Fatal("record snapshot followed current writes")
	}
	var old, other [768]float32
	if jointdecision.FeaturesIntoRecordThree(current.Context.Text, &old) != nil || jointdecision.FeaturesIntoRecordThree(saved.Context.Text, &other) != nil || old != other {
		t.Fatal("old model's known receiver collision changed")
	}
	if reflect.DeepEqual(current.ValueFlow, saved.ValueFlow) {
		t.Fatal("new origin graph still collides")
	}
}

func TestRecordFlowJoinsBranchWritesAndCarriesEarlyReturnCondition(t *testing.T) {
	body := "let copy = input0\ncopy.title = \"draft\"\nif input1 { copy.state = \"wait\" }\ncopy.reason = copy.state\nreturn copy"
	r := exportedFlow(t, flowSource(t, body)).ValueFlow
	if !flowContains(r, r.Choices[2].First, "join", "") || r.Choices[1].Guard == 0 || r.Choices[2].Guard != 0 {
		t.Fatal("conditional field definitions or joined continuation differ")
	}
	body = "let copy = input0\ncopy.title = \"draft\"\nif input1 { return copy }\ncopy.state = \"wait\"\ncopy.reason = copy.state\nreturn copy"
	r = exportedFlow(t, flowSource(t, body)).ValueFlow
	guard := r.Nodes[r.Choices[1].Guard-1]
	if guard.Kind != "guard" || guard.Operator != "false" || r.Choices[2].Guard != guard.ID {
		t.Fatal("continuation after early return lost its false branch")
	}
}

func TestRecordFlowRenamingAndCaseChangesPreserveOriginRelations(t *testing.T) {
	source := recordUpdatesFixture(t)
	a := exportedFlow(t, source).ValueFlow
	b := exportedFlow(t, []byte(strings.ReplaceAll(string(source), "copy", "work"))).ValueFlow
	a.BodySHA256, b.BodySHA256 = "", ""
	if !reflect.DeepEqual(a, b) {
		t.Fatal("equal-length local renaming changed origin relations")
	}
	changed := []byte(strings.Replace(string(source), `한글`, `새값`, 1))
	c := exportedFlow(t, changed).ValueFlow
	c.BodySHA256 = ""
	if !reflect.DeepEqual(a, c) {
		t.Fatal("finite input leaked into flow")
	}
	legacy, err := ExportRecordAssemblyContextWithFeature(context.Background(), "r.gooo", source, "Select", false, jointdecision.RecordSharedFeatureVersion)
	raw, _ := json.Marshal(legacy)
	if err != nil || legacy.ValueFlow != nil || strings.Contains(string(raw), "value_flow") {
		t.Fatal("legacy context gained optional observation")
	}
}

func TestRecordFlowScopeAndGraphAreDeterministic(t *testing.T) {
	body := "let copy = input0\ncopy.title = \"draft\"\nif input1 { let temp = \"unused\"; copy.state = temp }\ncopy.reason = copy.state\nreturn copy"
	source := flowSource(t, body)
	source = []byte(strings.Replace(string(source), `alternative "\"ready\""`, `alternative "temp + \"ready\""`, 1))
	r := exportedFlow(t, source).ValueFlow
	if len(r.Nodes) > recordFlowNodeLimit || len(r.Choices) != 3 {
		t.Fatal("flow bounds")
	}
	for i, node := range r.Nodes {
		if node.ID != uint16(i+1) || node.Parents[0] >= node.ID || node.Parents[1] >= node.ID || node.Guard >= node.ID {
			t.Fatal("definition graph must be ordered and acyclic", node)
		}
	}
	if !reflect.DeepEqual(r, exportedFlow(t, source).ValueFlow) {
		t.Fatal("repeated source observation differs")
	}
}
