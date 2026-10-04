package bodycodegen

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func assemblyFixture(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../examples/body-codegen/source-assembly.gooo.fixture")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestSourceAssemblyUsesExistingRecipePlanAndModelPipeline(t *testing.T) {
	ctx := context.Background()
	source := assemblyFixture(t)
	for _, activity := range []string{"Qualified", "Clamp"} {
		assembly, err := SourceAssembly(ctx, "inline.gooo", source, activity)
		if err != nil || assembly == nil {
			t.Fatal("source contract missing", err)
		}
		inline, err := DecodeSourcePathDocument(ctx, "inline.gooo", source, activity, nil)
		if err != nil {
			t.Fatal(err)
		}
		file, diagnostics := syntax.Parse(string(source))
		if diagnostics.HasErrors() {
			t.Fatal(diagnostics)
		}
		for _, declaration := range file.Declarations {
			if d, ok := declaration.(*syntax.ActivityDecl); ok {
				d.Assembly = nil
			}
		}
		plain, err := syntax.Format(file)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(sourceAssemblyRecipe(assembly))
		external, err := DecodeSourcePathDocument(ctx, "plain.gooo", []byte(plain), activity, raw)
		if err != nil || !reflect.DeepEqual(inline, external) {
			t.Fatal("inline contract changed the existing typed plan", err)
		}
		for _, model := range []string{"", writeTypedPathContractModel(t, false)} {
			result, err := GenerateWithTypedPaths(ctx, "inline.gooo", source, activity, inline, model)
			if err != nil {
				t.Fatal(err)
			}
			p := result.Report.BodyPaths
			if p.FunctionalCompleteness != 100 || !p.SourceBaseMatched || p.Search.Selection.ExternalCalls != 0 ||
				(model == "" && p.Search.Selection.ModelCalls != 0) || (model != "" && p.Search.Selection.ModelCalls != 3) {
				t.Fatal("assembly skipped the ordinary bounded pipeline", p)
			}
			if activity == "Qualified" && p.Search.Selection.Choices["base-local"] != "reference_second" {
				t.Fatal("non-default assembly was not selected")
			}
			if err := VerifyTypedPathProjection(ctx, "inline.gooo", source, inline, result); err != nil {
				t.Fatal("assembly replay failed", err)
			}
		}
	}
}

func TestSourceAssemblyContractOwnsCasesIntentAndBudgets(t *testing.T) {
	ctx := context.Background()
	source := assemblyFixture(t)
	document, err := DecodeSourcePathDocument(ctx, "inline.gooo", source, "Qualified", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeSourcePathDocument(ctx, "inline.gooo", source, "Qualified", []byte(`{}`)); err == nil {
		t.Fatal("external plan replaced source-owned assembly")
	}
	for _, mutate := range []func(*pathplan.Document){
		func(d *pathplan.Document) {
			d.TestCases = append([]pathplan.TestCase(nil), d.TestCases...)
			d.TestCases[0].Expected++
		},
		func(d *pathplan.Document) { d.MaxAttempts-- },
		func(d *pathplan.Document) {
			d.Plan.Decisions = append([]pathplan.Choice(nil), d.Plan.Decisions...)
			d.Plan.Decisions[0].Intent = "changed"
		},
		func(d *pathplan.Document) { d.Seed = "changed" },
	} {
		changed := document
		mutate(&changed)
		_, err := GenerateWithTypedPaths(ctx, "inline.gooo", source, "Qualified", changed, "missing-model.json")
		if err == nil || !strings.Contains(err.Error(), "source assembling contract") {
			t.Fatal("changed contract reached model loading", err)
		}
		if _, err = ExportTypedPathContext(ctx, "inline.gooo", source, "Qualified", changed); err == nil {
			t.Fatal("changed source contract reached context export")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = DecodeSourcePathDocument(cancelled, "inline.gooo", source, "Qualified", nil); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled assembly expanded", err)
	}
	if _, err = DecodeSourcePathDocument(ctx, "inline.gooo", recipeSource("return input"), "Assemble", nil); err == nil {
		t.Fatal("missing assembly silently invented a plan")
	}
}
