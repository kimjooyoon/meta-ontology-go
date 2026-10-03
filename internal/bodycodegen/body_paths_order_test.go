package bodycodegen

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/orderjudge"
	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
)

func orderFixture(t *testing.T) ([]byte, pathplan.Document) {
	t.Helper()
	source := recipeSource("let value = input; value = value + 1; value = value * 2; return value")
	raw := []byte(`{"schema":"gooo/source-typed-path-recipe/v1","choices":[
	{"id":"a","kind":"operand_order","occurrence":0,"intent":"Keep operands."},
	{"id":"b","kind":"operand_order","occurrence":1,"intent":"Keep operands."},
	{"id":"root","kind":"root_order","occurrence":1,"intent":"곱한 뒤 더한다. Multiply then add."}],
	"test_cases":[{"input":3,"expected":7},{"input":-1,"expected":-1}],"max_attempts":8}`)
	doc, err := DecodeSourcePathDocument(context.Background(), "order.gooo", source, "Assemble", raw)
	if err != nil {
		t.Fatal(err)
	}
	return source, doc
}

func writeOrderModel(t *testing.T) string {
	t.Helper()
	model, err := orderjudge.New([orderjudge.ParameterCount]float32{})
	if err != nil {
		t.Fatal(err)
	}
	meta, weights, err := model.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "weights.bin"), weights, 0600); err != nil {
		t.Fatal(err)
	}
	name := filepath.Join(dir, "model.json")
	if err := os.WriteFile(name, meta, 0600); err != nil {
		t.Fatal(err)
	}
	return name
}

func TestWholeCandidateModelRanksInsideCodegenAndReplays(t *testing.T) {
	source, doc := orderFixture(t)
	model := writeOrderModel(t)
	result, err := GenerateWithTypedPaths(context.Background(), "order.gooo", source, "Assemble", doc, model)
	if err != nil {
		t.Fatal(err)
	}
	p := result.Report.BodyPaths
	if !p.SourceBaseMatched || p.OrderJudgment == nil || p.Search.Selection.ModelCalls != 1 ||
		p.Search.Evaluated != 2 || len(p.OrderJudgment.Aliases) != 3 || p.FunctionalCompleteness != 100 ||
		p.ModelContext != nil || p.Search.Selection.ModelVariant != orderjudge.Schema || p.Search.Selection.ExternalCalls != 0 {
		t.Fatalf("whole-candidate path did not execute: %+v", p)
	}
	if err := VerifyTypedPathProjection(context.Background(), "order.gooo", source, doc, result); err != nil {
		t.Fatal(err)
	}
	disconnected, err := GenerateWithTypedPaths(context.Background(), "order.gooo", source, "Assemble", doc, "")
	if err != nil || disconnected.Report.BodyPaths.Search.Selection.ModelCalls != 0 || disconnected.Source != result.Source {
		t.Fatal("model-free result differs", err)
	}
	doc.MaxAttempts = 1
	partial, err := GenerateWithTypedPaths(context.Background(), "order.gooo", source, "Assemble", doc, model)
	if err != nil || partial.Report.BodyPaths.FunctionalCompleteness != 0 || partial.Report.BodyPaths.Search.Evaluated != 1 {
		t.Fatal("budget-one finite failure was hidden", err)
	}
	if err := VerifyTypedPathProjection(context.Background(), "order.gooo", source, doc, partial); err != nil {
		t.Fatal(err)
	}
}

func TestWholeCandidateSourceBindingAndUnsupportedOptionsPrecedePrediction(t *testing.T) {
	source, doc := orderFixture(t)
	changed := []byte(strings.Replace(string(source), "+ 1", "+ 2", 1))
	_, err := GenerateWithTypedPaths(context.Background(), "order.gooo", changed, "Assemble", doc, "missing-model.json")
	var failure *BodyPathError
	if !errors.As(err, &failure) || failure.Receipt.SearchStarted || failure.Receipt.SourceBaseMatched {
		t.Fatal("source mismatch reached the model", err)
	}
	model := writeOrderModel(t)
	for _, step := range []int{1, 8} {
		_, err = GenerateWithTypedPathBatches(context.Background(), "order.gooo", source, "Assemble", doc, model, step)
		if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 || !failure.Receipt.SourceBaseMatched {
			t.Fatal("unsupported batch was ignored", err)
		}
	}
	doc.Seed = "explicit-sample"
	_, err = GenerateWithTypedPaths(context.Background(), "order.gooo", source, "Assemble", doc, model)
	if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 {
		t.Fatal("unsupported seed was ignored", err)
	}
	doc.Seed = ""
	doc.Plan.Decisions[0].Intent = strings.Repeat("a", 512)
	_, err = GenerateWithTypedPaths(context.Background(), "order.gooo", source, "Assemble", doc, model)
	if !errors.As(err, &failure) || failure.Receipt.Search.Selection.ModelCalls != 0 || failure.Receipt.Search.Evaluated != 0 {
		t.Fatal("complete input was truncated or evaluated", err)
	}
}

func TestWholeCandidateRetainedConcurrencyAndBoundedArtifact(t *testing.T) {
	model := writeOrderModel(t)
	retained, err := NewTypedPathGenerator(model)
	if err != nil || retained.Info().ModelSchema != orderjudge.Schema || retained.Info().ResidentTensorBytes != 16384 {
		t.Fatal("retained model identity differs", err)
	}
	source, doc := orderFixture(t)
	var group sync.WaitGroup
	for range 4 {
		group.Go(func() {
			r, err := retained.Generate(context.Background(), "order.gooo", source, "Assemble", doc, TypedPathOptions{})
			if err != nil || r.Report.BodyPaths.Search.Selection.ModelCalls != 1 || r.Report.BodyPaths.FunctionalCompleteness != 100 {
				t.Error("parallel generation differs", err)
			}
		})
	}
	group.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := retained.Generate(ctx, "order.gooo", source, "Assemble", doc, TypedPathOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled generation ran", err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(model), "weights.bin"), make([]byte, 16385), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewTypedPathGenerator(model); err == nil {
		t.Fatal("oversized weights loaded")
	}
}
