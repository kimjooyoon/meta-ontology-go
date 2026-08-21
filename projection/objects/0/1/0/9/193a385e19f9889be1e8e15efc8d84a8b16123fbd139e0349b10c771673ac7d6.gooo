package pathclosure_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/provenance/pathclosure"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"reflect"
	"testing"
)

func TestEvaluatorRunsClosureOutcomePartitions(t *testing.T) {
	fixture := completeInferenceFixture()
	complete := requirementForPath("complete", fixture.edges)
	missing := complete
	missing.PathID = fixtureID("path/missing")
	missing.RecordIDs = append([]semantic.ID(nil), complete.RecordIDs...)
	missing.ExpectedKinds = append([]semantic.InferenceKind(nil), complete.ExpectedKinds...)
	missing.RecordIDs[1] = fixtureID("record/not-present")

	t.Run("complete", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{complete})
		assertEvaluation(t, got, pathclosure.PASS, pathclosure.CodePass, 1, 1)
		if !reflect.DeepEqual(got.Complete, []semantic.ID{complete.PathID}) {
			t.Fatalf("complete paths = %#v", got.Complete)
		}
	})
	t.Run("missing edge UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{missing})
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingRecord, 0, 1)
		if !reflect.DeepEqual(got.Missing, []semantic.ID{missing.PathID}) {
			t.Fatalf("missing paths = %#v", got.Missing)
		}
	})
	t.Run("zero requirements UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, nil)
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeZeroDenominator, 0, 0)
	})
	t.Run("two paths one incomplete UNKNOWN", func(t *testing.T) {
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{complete, missing})
		assertEvaluation(t, got, pathclosure.UNKNOWN, pathclosure.CodeMissingRecord, 1, 2)
		if !reflect.DeepEqual(got.Complete, []semantic.ID{complete.PathID}) ||
			!reflect.DeepEqual(got.Missing, []semantic.ID{missing.PathID}) {
			t.Fatalf("mixed paths = %#v", got)
		}
	})
	t.Run("wrong phase order FAIL_CLOSED", func(t *testing.T) {
		wrongOrder := requirementForPath("wrong-order", []semantic.InferenceEdge{fixture.edges[1], fixture.edges[0]})
		got := pathclosure.Evaluate(fixture.path, []pathclosure.Requirement{wrongOrder})
		assertEvaluation(t, got, pathclosure.FAIL_CLOSED, pathclosure.CodeMalformed, 0, 1)
		if !reflect.DeepEqual(got.Malformed, []semantic.ID{wrongOrder.PathID}) {
			t.Fatalf("malformed paths = %#v", got.Malformed)
		}
	})
}
