package query

import (
	"errors"
	"testing"
)

func TestInferenceProjectionBoundariesNeverPassPartials(t *testing.T) {
	path, _ := inferenceQueryFixture(t)
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	exact := inferenceQueryRequest()
	exact.Limit = 6
	exact.MaxWork = 6
	exactResult, err := projection.Execute(exact)
	if err != nil || !exactResult.Complete || exactResult.Work.Used != 6 {
		t.Fatalf("exact work boundary = %#v err=%v", exactResult, err)
	}
	oneOver := exact
	oneOver.MaxWork = 5
	overResult, err := projection.Execute(oneOver)
	if !errors.Is(err, ErrInferenceQueryBudget) || overResult.Complete || len(overResult.Edges) != 0 || overResult.Work.Used != 5 {
		t.Fatalf("one-over work boundary = %#v err=%v", overResult, err)
	}
	rowOver := exact
	rowOver.MaxWork = 32
	rowOver.Limit = 5
	rowResult, err := projection.Execute(rowOver)
	if !errors.Is(err, ErrInferenceQueryBudget) || rowResult.Complete || len(rowResult.Edges) != 0 {
		t.Fatalf("row overrun = %#v err=%v", rowResult, err)
	}
	unsupported := exact
	unsupported.Predicate = "not-a-semantic-predicate"
	badPredicate, err := projection.Execute(unsupported)
	if !errors.Is(err, ErrInferenceUnsupportedPred) || badPredicate.Complete {
		t.Fatalf("unsupported predicate = %#v err=%v", badPredicate, err)
	}
	depth := exact
	depth.Explain = true
	depth.MaxDepth = 5
	depthResult, err := projection.Execute(depth)
	if !errors.Is(err, ErrInferenceQueryBudget) || depthResult.Complete || len(depthResult.Edges) != 0 {
		t.Fatalf("depth overrun = %#v err=%v", depthResult, err)
	}
}
