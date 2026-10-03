package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestRecipeProjectionReuseKeepsFallbackBinding(t *testing.T) {
	source := recipeSource("return input - 2")
	ctx := context.Background()
	doc, err := DecodeSourcePathDocument(ctx, "source.gooo", source, "Assemble", recipeBytes(recipeOperand))
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := doc.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	base, err := GenerateWithPlanner(ctx, "source.gooo", source, "Assemble", "", "")
	if err != nil {
		t.Fatal(err)
	}
	freshReceipt, reusedReceipt := &BodyPathReceipt{}, &BodyPathReceipt{}
	fresh, err := bindTypedPathSource(ctx, "source.gooo", source, "Assemble", prepared, freshReceipt)
	if err != nil {
		t.Fatal(err)
	}
	reused, err := bindTypedPathSourceProjection(ctx, "source.gooo", source, "Assemble", prepared, reusedReceipt, &base)
	if err != nil || fresh.base.Source != reused.base.Source ||
		!reflect.DeepEqual(freshReceipt.SourceBinding, reusedReceipt.SourceBinding) || !reusedReceipt.SourceBaseMatched {
		t.Fatal("source equivalence changed", err)
	}
	changed := []byte(strings.Replace(string(source), "input - 2", "input - 3", 1))
	changedBase, err := GenerateWithPlanner(ctx, "source.gooo", changed, "Assemble", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bindTypedPathSourceProjection(ctx, "source.gooo", changed, "Assemble", prepared,
		&BodyPathReceipt{}, &changedBase); err == nil || !strings.Contains(err.Error(), "fallback does not match") {
		t.Fatal("a current projection bypassed fallback binding", err)
	}
	if _, err := bindTypedPathSourceProjection(ctx, "source.gooo", changed, "Assemble", prepared,
		&BodyPathReceipt{}, &base); err == nil || !strings.Contains(err.Error(), "current checked source") {
		t.Fatal("a stale projection was accepted", err)
	}
}

func TestRecipeProjectionRejectsAlteredEvidenceAndCancellation(t *testing.T) {
	source := recipeSource("return input - 2")
	base, err := Generate("source.gooo", source, "Assemble")
	if err != nil {
		t.Fatal(err)
	}
	for _, alter := range []func(*Result){
		func(r *Result) { r.Report.Activity = "Other" },
		func(r *Result) { r.Report.SourceDigest = "stale" },
		func(r *Result) { r.Source += "\n// changed" },
		func(r *Result) { r.Report.Decision = "FAIL_CLOSED" },
		func(r *Result) { r.Report.TypecheckPassed = false },
		func(r *Result) { r.Report.DeterministicReplay = false },
		func(r *Result) { r.Report.RepositoryWrites = 1 },
	} {
		changed := base
		alter(&changed)
		if _, err := typedPathBindingProjection(context.Background(), "source.gooo", source, "Assemble", &changed); err == nil {
			t.Fatal("altered projection accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := typedPathBindingProjection(ctx, "source.gooo", source, "Assemble", &base); err == nil {
		t.Fatal("cancelled projection reuse accepted")
	}
}
