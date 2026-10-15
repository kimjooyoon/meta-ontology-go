package bodyexecution

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestCriteriaQuestionsCallExposureIncludesUnselectedOutputs(t *testing.T) {
	graph, seen := localInputFixture()
	seen.unobservedConstructionCalls = false
	partial := localInputObservation("seed", "[17]")
	partial.Status = "UNKNOWN"
	questions := &bodycodegen.CriteriaQuestionReceipt{
		Outputs: []bodycodegen.CriteriaQuestionOutput{{Input: "[7]", LocalCalls: localInputObservation("seed", "[10]")},
			{Input: "[-5]", LocalCalls: partial}},
		Evaluations: []bodycodegen.CriteriaQuestionEvaluation{{Result: bodycodegen.RecordCriteriaInput{Checks: []bodycodegen.RecordCriterionCheck{
			{LocalCalls: localInputObservation("seed", "[23]")},
		}}}},
	}
	prior := Composition{Steps: []CompositionStep{{Generation: bodycodegen.Result{Report: bodycodegen.Report{
		ActivityID: "root", RecordAssembly: &bodycodegen.RecordAssemblyReceipt{CriteriaQuestions: questions},
	}}}}}
	if err := graph.includeCompositionLocalInputs(&seen, prior); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"[10]", "[17]", "[23]"} {
		if _, ok := seen.called["seed"][input]; !ok {
			t.Fatal("question input exposure disappeared", input)
		}
	}
	if !seen.unobservedConstructionCalls {
		t.Fatal("partial capture claimed absence evidence")
	}
}
