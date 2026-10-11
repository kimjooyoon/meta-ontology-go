package bodyexecution

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func localInputFixture() (compositionGraph, compositionSelectionInputs) {
	graph := compositionGraph{count: 1,
		called: map[string]CompositionActivity{"seed": {ID: "seed", Name: "Seed", InputType: "Integer"}}}
	graph.nodes[0] = CompositionActivity{ID: "root", Name: "Root", InputType: "Integer", Assembling: true}
	seen := compositionSelectionInputs{called: map[string]map[string]struct{}{"seed": {}},
		unobservedConstructionCalls: true}
	seen.nodes[0] = map[string]struct{}{}
	return graph, seen
}

func localInputObservation(target, inputs string) *bodycodegen.LocalCallObservation {
	return &bodycodegen.LocalCallObservation{Schema: "gooo/local-call-inputs/v1", ActivityID: "root", Status: "RECORDED",
		Calls: []bodycodegen.LocalCallInput{{ActivityID: target, Inputs: json.RawMessage(inputs)}}}
}

func TestCompositionLocalInputsIncludeFailedAndPartialPrefixes(t *testing.T) {
	graph, seen := localInputFixture()
	failed := localInputObservation("seed", "[9007199254740993]")
	failed.EvaluationError = "zero divisor"
	partial := localInputObservation("root", "[4]")
	partial.Status, partial.Reason = "UNKNOWN", "CALL_RECORD_LIMIT"
	prior := Composition{Preparations: []CompositionStep{{Generation: bodycodegen.Result{Report: bodycodegen.Report{
		ActivityID: "root", BodySearch: &bodycodegen.IRBodySearchReceipt{LocalCalls: &bodycodegen.LocalCallHistory{
			Schema: "gooo/local-call-inputs/v1", Status: "RECORDED", Evaluations: []*bodycodegen.LocalCallObservation{failed}}},
	}}}}, Steps: []CompositionStep{{Generation: bodycodegen.Result{Report: bodycodegen.Report{
		ActivityID: "root", RecordAssembly: &bodycodegen.RecordAssemblyReceipt{Attempts: []bodycodegen.RecordAssemblyAttempt{
			{Status: "EVALUATION_FAILED", LocalCalls: partial},
		}},
	}}}}}
	if err := graph.includeCompositionLocalInputs(&seen, prior); err != nil {
		t.Fatal(err)
	}
	if _, ok := seen.called["seed"]["[9007199254740993]"]; !ok {
		t.Fatal("failed helper input lost precision or disappeared")
	}
	if _, ok := seen.nodes[0]["[4]"]; !ok || !seen.unobservedConstructionCalls {
		t.Fatal("partial prefix became unobserved or gained absence evidence")
	}
}

func TestJointLocalInputsIncludeRejectedCandidatePrefix(t *testing.T) {
	graph, seen := localInputFixture()
	attempt := JointAttempt{Rejection: &JointCandidateRejection{Stage: "LOCAL_SOURCE_SEARCH"},
		Candidates: []bodycodegen.RecordCandidate{{ActivityID: "root", Attempt: bodycodegen.RecordAssemblyAttempt{
			LocalCalls: localInputObservation("seed", "[8]")}}},
		SearchCandidates: []bodycodegen.SearchCandidate{{ActivityID: "root", LocalCalls: &bodycodegen.LocalCallHistory{
			Schema: "gooo/local-call-inputs/v1", Status: "RECORDED",
			Evaluations: []*bodycodegen.LocalCallObservation{localInputObservation("seed", "[9]")},
		}}},
	}
	if err := graph.includeJointLocalInputs(&seen, attempt); err != nil {
		t.Fatal(err)
	}
	if len(seen.called["seed"]) != 2 || !seen.unobservedConstructionCalls || attempt.Runtime.Stage != "" {
		t.Fatal("rejected prefix was lost or gained native execution evidence")
	}
}

func TestLocalInputEvidencePreservesUnknownAndRejectsMalformedBindings(t *testing.T) {
	graph, seen := localInputFixture()
	if err := graph.includeLocalCallHistory(&seen, "root", nil); err != nil || !seen.unobservedConstructionCalls {
		t.Fatal("historical absence was upgraded", err)
	}
	seen.unobservedConstructionCalls = false
	partial := localInputObservation("seed", "[5]")
	partial.Status, partial.Reason = "UNKNOWN", "CALL_INPUT_BYTE_LIMIT"
	if err := graph.includeLocalCallObservation(&seen, "root", partial); err != nil || !seen.unobservedConstructionCalls {
		t.Fatal("bounded history claimed complete absence evidence", err)
	}
	for _, tc := range []struct{ activity, target, inputs string }{
		{"wrong-root", "seed", "[1]"}, {"root", "seed", "[1,2]"},
		{"root", "seed", "1"}, {"root", "seed", `["wrong-type"]`},
	} {
		if err := graph.includeLocalCallObservation(&seen, tc.activity, localInputObservation(tc.target, tc.inputs)); err == nil {
			t.Fatal("malformed typed input evidence accepted", tc)
		}
	}
	if err := graph.includeLocalCallObservation(&seen, "root", localInputObservation("fixed-helper", "[1]")); err != nil {
		t.Fatal("fixed helper invented a constructed target", err)
	}
}
