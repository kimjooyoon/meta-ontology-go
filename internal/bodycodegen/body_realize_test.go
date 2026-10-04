package bodycodegen

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestSourceAssemblyBodyOnlyAdoptionRequiresRebase(t *testing.T) {
	ctx, source := context.Background(), assemblyFixture(t)
	document, err := DecodeSourcePathDocument(ctx, "original.gooo", source, "Qualified", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "original.gooo", source, "Qualified", document, "")
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := document.Prepare()
	selected, err := prepared.Compile(result.Report.BodyPaths.Search.Selection.Choices)
	if err != nil {
		t.Fatal(err)
	}
	file, _ := syntax.ParseFile("original.gooo", string(source))
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if d, ok := declaration.(*syntax.ActivityDecl); ok && d.Name == "Qualified" {
			activity = d
		}
	}
	changed, err := replaceActivityProgram(source, activity.ValueProgramSpan, selected.GoooBody())
	if err != nil || digest(changed) == result.Report.BodyPaths.SelectedSourceSHA256 {
		t.Fatal("selected source reconstruction differs", err)
	}
	_, err = DecodeSourcePathDocument(ctx, "changed.gooo", changed, "Qualified", nil)
	if err == nil || !strings.Contains(err.Error(), "distinct alternative") {
		t.Fatal("body-only adoption did not expose the stale source selector", err)
	}
}

func TestSourceAssemblyRealizationIsReusableAndLocal(t *testing.T) {
	ctx, source := context.Background(), assemblyFixture(t)
	document, err := DecodeSourcePathDocument(ctx, "original.gooo", source, "Qualified", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "original.gooo", source, "Qualified", document, "")
	if err != nil {
		t.Fatal(err)
	}
	realized, err := RealizeSourceAssembly(ctx, "original.gooo", source, result)
	if err != nil || realized.ModelCalls != 0 || realized.Source != result.GoooSource ||
		realized.FinitePassed != 5 || realized.FiniteTotal != 5 || realized.LegacyCheckpointUpgrade {
		t.Fatal("realization differs", err, realized)
	}
	again, err := DecodeSourcePathDocument(ctx, "realized.gooo", []byte(realized.Source), "Qualified", nil)
	if err != nil || !reflect.DeepEqual(document, again) {
		t.Fatal("checkpoint changed original alternatives", err)
	}
	next, err := GenerateWithTypedPaths(ctx, "realized.gooo", []byte(realized.Source), "Qualified", again, "")
	if err != nil || next.Source != result.Source || next.GoooSource != realized.Source ||
		next.Report.ActivityID != result.Report.ActivityID {
		t.Fatal("realized generation is not a fixed point", err)
	}
	if err := VerifyTypedPathProjection(ctx, "realized.gooo", []byte(realized.Source), again, next); err != nil {
		t.Fatal(err)
	}
	marker := "\nactivity Clamp"
	if string(source[strings.Index(string(source), marker):]) != realized.Source[strings.Index(realized.Source, marker):] {
		t.Fatal("realization changed an unrelated activity")
	}
	changed := result
	changed.GoooSource += "\n"
	if _, err := RealizeSourceAssembly(ctx, "original.gooo", source, changed); err == nil {
		t.Fatal("mismatched source projection was realized")
	}
}

func TestSourceAssemblyLegacyRealizationRetainsOriginalChoices(t *testing.T) {
	ctx, source := context.Background(), assemblyFixture(t)
	document, err := DecodeSourcePathDocument(ctx, "original.gooo", source, "Qualified", nil)
	if err != nil {
		t.Fatal(err)
	}
	result, err := GenerateWithTypedPaths(ctx, "original.gooo", source, "Qualified", document, "")
	if err != nil {
		t.Fatal(err)
	}
	prepared, _ := document.Prepare()
	selected, _ := prepared.Compile(bodyPathSelection(result.Report.BodyPaths).Choices)
	file, _ := syntax.Parse(string(source))
	activity := file.Declarations[1].(*syntax.ActivityDecl)
	bodyOnly, err := replaceActivityProgram(source, activity.ValueProgramSpan, selected.GoooBody())
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := GenerateWithPlanner(ctx, "original.gooo", bodyOnly, "Qualified", "", "")
	if err != nil {
		t.Fatal(err)
	}
	receipt := *result.Report.BodyPaths
	receipt.SourceFormat, receipt.SelectedSourceSHA256 = "", digest(bodyOnly)
	legacy.Report.BodyPaths = &receipt
	populateCompletenessReceipt(&legacy.Report, "")
	realized, err := RealizeSourceAssembly(ctx, "original.gooo", source, legacy)
	if err != nil || !realized.LegacyCheckpointUpgrade || realized.Source != result.GoooSource ||
		realized.GenerationSourceSHA256 == realized.RealizedSourceSHA256 {
		t.Fatal("legacy upgrade lost its origin", err, realized)
	}
}

func TestSourceAssemblyCheckpointMismatchPrecedesModelLoading(t *testing.T) {
	ctx, source := context.Background(), assemblyFixture(t)
	document, _ := DecodeSourcePathDocument(ctx, "original.gooo", source, "Qualified", nil)
	result, err := GenerateWithTypedPaths(ctx, "original.gooo", source, "Qualified", document, "")
	if err != nil {
		t.Fatal(err)
	}
	file, _ := syntax.Parse(result.GoooSource)
	activity := file.Declarations[1].(*syntax.ActivityDecl)
	changed, err := replaceActivityProgram([]byte(result.GoooSource), activity.ValueProgramSpan, "return input")
	if err != nil {
		t.Fatal(err)
	}
	_, err = GenerateWithTypedPaths(ctx, "changed.gooo", changed, "Qualified", document, "missing-model.json")
	if err == nil || !strings.Contains(err.Error(), "computes body differs") {
		t.Fatal("checkpoint inconsistency reached model loading", err)
	}
}
