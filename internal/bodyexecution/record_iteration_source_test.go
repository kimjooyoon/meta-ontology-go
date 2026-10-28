package bodyexecution

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func TestRecordIterationFieldSourceOrigins(t *testing.T) {
	source, err := os.ReadFile("../../examples/record-collections/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("../../examples/record-collections/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "items.gooo", source, suite, CompositionOptions{EntryActivity: "Total"})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := ExecuteComposition(context.Background(), "items.gooo", source, plan, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if runtime.FinitePassed != 4 || runtime.FiniteTotal != 4 || !runtime.RuntimeReplayed || runtime.ModelCalls != 0 {
		t.Fatal("lost existing exact totals or replay")
	}
	if runtime.FieldSources.Schema != "gooo/native-field-sources/v2" {
		t.Fatal("derived source table kept a legacy schema")
	}
	mapped := 0
	for _, binding := range runtime.FieldSources.Bindings {
		if binding.Kind != "write" {
			continue
		}
		if binding.Source.Value.Status != "MAPPED" || binding.Source.Target.Status != "MAPPED" {
			t.Fatal("range-derived field write lost its actual source value/target", binding.SiteIndex, binding.Source.Value.Status, binding.Source.Target.Status)
		}
		if binding.Origin != nil {
			t.Fatal("range binding was relabelled as an authored field assignment")
		}
		if binding.Source.Value.Expression != "input.items" || binding.Source.Target.Expression != "item" {
			t.Fatal("invented field assignment expression", binding.Source)
		}
		for _, span := range []*bodycodegen.BodyExpressionSource{binding.Source.Value, binding.Source.Target} {
			if span.Start < 0 || span.End > len(plan.GoooSource) || span.SourceSHA256 != digest([]byte(plan.GoooSource)) || plan.GoooSource[span.Start:span.End] != span.Expression {
				t.Fatal("range source does not name exact original bytes", span)
			}
		}
		if strings.Contains(binding.Source.Target.Expression, ".") {
			t.Fatal("derived field syntax was presented as authored")
		}
		if binding.Derivation == nil || binding.Derivation.Kind != "record_iteration_field" || binding.Derivation.Local.Kind != "iteration" {
			t.Fatal("missing explicit range derivation", binding)
		}
		mapped++
	}
	if mapped != 2 {
		t.Fatal("derived field population changed", mapped)
	}
	if err := validateFieldSources(runtime); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*CompositionFieldSourceBinding){
		func(b *CompositionFieldSourceBinding) { b.Derivation.FieldID = "wrong-field" },
		func(b *CompositionFieldSourceBinding) { b.Derivation.Local.VariableOrdinal++ },
		func(b *CompositionFieldSourceBinding) { b.Derivation.Local.Kind = "assignment" },
		func(b *CompositionFieldSourceBinding) { b.Derivation.Local.Name = "other" },
		func(b *CompositionFieldSourceBinding) { b.Derivation.Local.Operation = "=" },
		func(b *CompositionFieldSourceBinding) { b.Derivation = nil },
	} {
		encoded, _ := json.Marshal(runtime)
		var altered CompositionRuntime
		if err := json.Unmarshal(encoded, &altered); err != nil {
			t.Fatal(err)
		}
		change(&altered.FieldSources.Bindings[0])
		altered.FieldSources.SHA256 = fieldSourcesDigest(*altered.FieldSources)
		if err := validateFieldSources(altered); err == nil {
			t.Fatal("contradictory range origin accepted after checksum update")
		}
	}
}

func TestRecordIterationFieldSourceDisplayRename(t *testing.T) {
	source, err := os.ReadFile("../../examples/record-collections/main.gooo")
	if err != nil {
		t.Fatal(err)
	}
	source = []byte(strings.ReplaceAll(strings.ReplaceAll(string(source), "for item in", "for row in"), "item.", "row."))
	raw, err := os.ReadFile("../../examples/record-collections/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := DecodeCompositionCases(raw)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "relocated.gooo", source, suite, CompositionOptions{EntryActivity: "Total"})
	if err != nil {
		t.Fatal(err)
	}
	observed, err := ExecuteComposition(context.Background(), "relocated.gooo", source, plan, suite, "")
	if err != nil {
		t.Fatal(err)
	}
	if observed.FinitePassed != 4 || !observed.RuntimeReplayed || observed.ModelCalls != 0 {
		t.Fatal("rename changed values or replay")
	}
	count := 0
	for _, binding := range observed.FieldSources.Bindings {
		if binding.Kind != "write" {
			continue
		}
		if binding.Derivation == nil || binding.Source.Value.Expression != "input.items" || binding.Source.Target.Expression != "row" || binding.Source.Value.Status != "MAPPED" || binding.Source.Target.Status != "MAPPED" {
			t.Fatal("display rename lost original range source", binding)
		}
		if binding.Derivation.RecordID != "urn:record-items:line" {
			t.Fatal("record identity changed")
		}
		count++
	}
	if count != 2 {
		t.Fatal("renamed derivation population", count)
	}
}

func TestRecordIterationAbsentLocalAnnotations(t *testing.T) {
	source, cases := recordIterationExample(t)
	suite, err := DecodeCompositionCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := GenerateCompositionWithOptions(context.Background(), "historical.gooo", source, suite, CompositionOptions{EntryActivity: "Total"})
	if err != nil {
		t.Fatal(err)
	}
	for i := range plan.Steps {
		plan.Steps[i].Generation.Report.BodyLocals = nil
	}
	r, err := ExecuteComposition(context.Background(), "historical.gooo", source, plan, suite, "")
	if err != nil || r.FinitePassed != 4 || !r.RuntimeReplayed || r.ModelCalls != 0 {
		t.Fatal("missing historical source annotation interrupted computation", err, r.FinitePassed)
	}
	if r.BridgeObservation != nil &&
		(r.BridgeObservation.ElementLinks != 0 || r.BridgeObservation.Schema != "gooo/native-field-local-bridge-summary/v1") {
		t.Fatal("invented an absent source binding", r.BridgeObservation)
	}
	t.Log("4 original computations and replay survive absent historical local annotations;no new element-source claim invented;models0")
}
