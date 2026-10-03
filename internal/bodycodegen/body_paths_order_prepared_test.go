package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func TestWholeCandidatePreparedReuseRevalidatesCasesAndSource(t *testing.T) {
	source, doc := orderFixture(t)
	g, err := NewTypedPathGenerator(writeOrderModel(t))
	if err != nil {
		t.Fatal(err)
	}
	generate := func(source []byte, doc pathplan.Document) Result {
		t.Helper()
		r, err := g.Generate(context.Background(), "order.gooo", source, "Assemble", doc, TypedPathOptions{})
		if err != nil {
			t.Fatal(err)
		}
		p := r.Report.BodyPaths
		if p.OrderPreparation == nil || p.OrderPreparation.CandidateCount != 8 ||
			p.OrderPreparation.PlanSHA256 != p.Search.Selection.PlanSHA256 || p.Search.Selection.ModelCalls != 1 || !p.SourceBaseMatched {
			t.Fatal("prepared observation missing", p)
		}
		if err := VerifyTypedPathProjection(context.Background(), "order.gooo", source, doc, r); err != nil {
			t.Fatal(err)
		}
		return r
	}
	first, next := generate(source, doc), generate(source, doc)
	if first.Report.BodyPaths.OrderPreparation.Reused || !next.Report.BodyPaths.OrderPreparation.Reused ||
		first.Source != next.Source || !reflect.DeepEqual(first.Report.BodyPaths.Search, next.Report.BodyPaths.Search) ||
		first.Report.BodyPaths.OrderJudgment.Prediction != next.Report.BodyPaths.OrderJudgment.Prediction {
		t.Fatal("retained selection differs")
	}
	differentCases := doc
	differentCases.TestCases = []pathplan.TestCase{{Input: 3, Expected: 999}}
	partial := generate(source, differentCases)
	if !partial.Report.BodyPaths.OrderPreparation.Reused || partial.Report.BodyPaths.FunctionalCompleteness != 0 ||
		partial.Report.BodyPaths.Search.Status != "PARTIAL" {
		t.Fatal("a previous answer was reused as correctness")
	}
	changed := []byte(strings.Replace(string(source), "+ 1", "+ 2", 1))
	if _, err := g.Generate(context.Background(), "order.gooo", changed, "Assemble", doc, TypedPathOptions{}); err == nil {
		t.Fatal("cache bypassed source binding")
	}
	var updated pathplan.Document
	raw, _ := json.Marshal(doc)
	if err := json.Unmarshal(raw, &updated); err != nil {
		t.Fatal(err)
	}
	for i := range updated.Plan.Base.Expressions {
		e := &updated.Plan.Base.Expressions[i]
		if e.Kind == "int" && e.Int == 1 {
			e.Int = 2
		}
	}
	updated.TestCases = []pathplan.TestCase{{Input: 3, Expected: 8}, {Input: -1, Expected: 0}}
	fresh := generate(changed, updated)
	if fresh.Report.BodyPaths.OrderPreparation.Reused || fresh.Report.BodyPaths.FunctionalCompleteness != 100 {
		t.Fatal("changed plan reused stale programs")
	}
	back := generate(source, doc)
	if back.Report.BodyPaths.OrderPreparation.Reused || back.Source != first.Source {
		t.Fatal("single retained plan did not replace its predecessor")
	}
	var callers sync.WaitGroup
	for i := range 4 {
		callers.Go(func() {
			callSource, callDoc := source, doc
			if i%2 == 1 {
				callSource, callDoc = changed, updated
			}
			for range 3 {
				r, err := g.Generate(context.Background(), "order.gooo", callSource, "Assemble", callDoc, TypedPathOptions{})
				if err != nil || r.Report.BodyPaths.Search.Selection.ModelCalls != 1 || r.Report.BodyPaths.FunctionalCompleteness != 100 {
					t.Error("concurrent source replacement changed the result", err)
				}
			}
		})
	}
	callers.Wait()
}
