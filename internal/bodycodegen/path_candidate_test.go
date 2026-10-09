package bodycodegen

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/gooo-decision-runtime/pathplan"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestSourcePathCandidatesIncludeInitialWithinSourceBudget(t *testing.T) {
	source := checkpointFixture(t, "let a = input - 1; let b = a - 3; return b - 9", []assemblyspec.Choice{
		{ID: "a", Kind: "operand_order", Occurrence: 0, Intent: "First operand."},
		{ID: "b", Kind: "operand_order", Occurrence: 1, Intent: "Second operand."},
		{ID: "c", Kind: "operand_order", Occurrence: 2, Intent: "Third operand."},
	})
	source = []byte(strings.Replace(string(source), `attempts "64"`, `attempts "2"`, 1))
	ctx := context.Background()
	document, err := DecodeSourcePathDocument(ctx, "paths.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := document.Prepare()
	raw, _ := json.Marshal(document)
	// Model selection may be outside the deterministic two-row prefix. Its
	// historical selected mask still consumes a source attempt; no new ranking.
	prior := &BodyPathReceipt{DocumentSHA256: digest(raw), Search: pathplan.SearchResult{
		Selection: pathplan.Selection{PlanSHA256: prepared.PlanSHA256(),
			Choices: map[string]string{"a": "layout_reverse", "b": "layout_reverse", "c": "layout_reverse"}}}}
	set, err := PlanSourcePathCandidates(ctx, "paths.gooo", source, "Assemble", prior)
	if err != nil || set.AttemptBudget != 2 || set.Declared != 8 || !reflect.DeepEqual(set.Masks, []uint16{7, 0}) {
		t.Fatal(set, err)
	}
	if _, _, err := RealizeSourcePathCandidate(ctx, "paths.gooo", source, "Assemble", prior, 1); err == nil {
		t.Fatal("direct realization escaped the source budget")
	}
	selected, completed, err := RealizeSourcePathCandidate(ctx, "paths.gooo", source, "Assemble", prior, 7)
	if err != nil || selected.Total != 1 || selected.Stage != "COMPLETE" || len(completed) == 0 ||
		strings.Contains(string(completed), "assembling") {
		t.Fatal(selected, err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := PlanSourcePathCandidates(canceled, "paths.gooo", source, "Assemble", prior); err == nil {
		t.Fatal("canceled candidate planning succeeded")
	}
	prior.DocumentSHA256 = "changed"
	if _, err := PlanSourcePathCandidates(ctx, "paths.gooo", source, "Assemble", prior); err == nil {
		t.Fatal("unbound initial document succeeded")
	}
}

func TestSourcePathCandidatesRetainEncodedInitialModelPlan(t *testing.T) {
	source := checkpointFixture(t, "return input - 2", []assemblyspec.Choice{
		{ID: "offset", Kind: "operand_order", Intent: "Reverse the operands."},
	})
	ctx := context.Background()
	document, err := DecodeSourcePathDocument(ctx, "encoded.gooo", source, "Assemble", nil)
	if err != nil {
		t.Fatal(err)
	}
	prior, err := GenerateWithTypedPaths(ctx, "encoded.gooo", source, "Assemble", document, writePathContextContractModel(t))
	if err != nil || prior.Report.BodyPaths.ModelContext == nil || prior.Report.BodyPaths.ModelContext.Status != "ENCODED" {
		t.Fatal("initial encoded model context missing", err)
	}
	paths := prior.Report.BodyPaths
	set, err := PlanSourcePathCandidates(ctx, "encoded.gooo", source, "Assemble", paths)
	if err != nil || len(set.Masks) != 2 || set.PlanSHA256 == paths.Search.Selection.PlanSHA256 || paths.Search.Selection.ModelCalls != 1 {
		t.Fatal("encoded ranking was treated as an original source plan", set, err)
	}
	if _, _, err := RealizeSourcePathCandidate(ctx, "encoded.gooo", source, "Assemble", paths, set.Masks[0]); err != nil {
		t.Fatal(err)
	}
}
