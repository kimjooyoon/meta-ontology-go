package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestFixtureLocalityKeepsUnchangedRegionsByteStable(t *testing.T) {
	first := mustAcceptanceResult(t, acceptanceFixture(), nil)
	previous := bytes.Replace(first.Source, []byte("package bootstrapgen\n"), []byte("package bootstrapgen\n\nvar Keep = 7\n"), 1)
	changed := acceptanceFixture()
	changed.Activities[0].Name = "CompileBootstrap"
	changed.Activities[0].GoName = "CompileBootstrap"
	changed.Activities = append(changed.Activities, Activity{
		ID: "gooo://activity/publish", Name: "Publish", GoName: "Publish",
		Inputs: []Port{{EntityID: "gooo://entity/artifact", Name: "artifact"}},
	})
	second := mustAcceptanceResult(t, changed, previous)
	if !strings.Contains(string(second.Source), "var Keep = 7") {
		t.Fatal("marker-outside handwritten text was not preserved")
	}
	for _, id := range []string{"gooo://entity/source", "gooo://activity/inspect"} {
		if !bytes.Equal(testGeneratedBlock(t, first.Source, id), testGeneratedBlock(t, second.Source, id)) {
			t.Fatalf("unchanged generated region %q changed", id)
		}
	}
	if !strings.Contains(string(second.Source), `//gooo:generated:start id="gooo://activity/publish"`) {
		t.Fatal("new generated region was not appended")
	}
}

func TestFixtureCorruptionRejectsWithoutChangingPrevious(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corruptions := []string{
		strings.Replace(string(result.Source), `//gooo:generated:end id="gooo://activity/compile"`, `//gooo:generated:end id="wrong"`, 1),
		strings.Replace(string(result.Source), `//gooo:slot:end id="gooo://slot/compile-implementation"`, `//gooo:slot:end id="wrong"`, 1),
	}
	for index, corrupted := range corruptions {
		previous := []byte(corrupted)
		if _, err := Generate(acceptanceFixture(), previous); err == nil {
			t.Fatalf("corruption %d was accepted", index)
		}
		if !bytes.Equal(previous, []byte(corrupted)) {
			t.Fatalf("corruption %d changed caller-owned previous source", index)
		}
	}
}

func TestFixtureStrictMarkersRejectUnknownAttributesAndKindMismatch(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corruptions := []string{
		strings.Replace(string(result.Source), `kind="entity"`, `kind="entity" extra="x"`, 1),
		strings.Replace(string(result.Source), `//gooo:generated:end id="gooo://activity/compile" kind="activity"`, `//gooo:generated:end id="gooo://activity/compile" kind="entity"`, 1),
		strings.Replace(string(result.Source), `//gooo:generated:end id="gooo://activity/compile" kind="activity"`, `//gooo:generated:end kind="activity"`, 1),
		strings.Replace(string(result.Source), `//gooo:slot:start id="gooo://slot/compile-implementation"`, `//gooo:slot:start id="gooo://slot/compile-implementation" extra="x"`, 1),
	}
	for index, corrupted := range corruptions {
		previous := []byte(corrupted)
		_, err := Generate(acceptanceFixture(), previous)
		if err == nil {
			t.Fatalf("corruption %d was accepted", index)
		}
		if !bytes.Equal(previous, []byte(corrupted)) {
			t.Fatalf("corruption %d changed caller-owned previous source", index)
		}
	}
}

func TestFixtureRegionAndSlotStableIDCollisionRejectsWithoutMutation(t *testing.T) {
	result := mustAcceptanceResult(t, acceptanceFixture(), nil)
	corrupted := strings.Replace(string(result.Source), `//gooo:slot:start id="gooo://slot/compile-implementation"`, `//gooo:slot:start id="gooo://entity/source"`, 1)
	corrupted = strings.Replace(corrupted, `//gooo:slot:end id="gooo://slot/compile-implementation"`, `//gooo:slot:end id="gooo://entity/source"`, 1)
	previous := []byte(corrupted)
	_, err := Generate(acceptanceFixture(), previous)
	if err == nil || !strings.Contains(err.Error(), "stable ID") {
		t.Fatalf("expected stable-ID collision rejection, got %v", err)
	}
	if !bytes.Equal(previous, []byte(corrupted)) {
		t.Fatal("stable-ID rejection changed caller-owned previous source")
	}
}

func TestFixtureSourceMapKeepsStableSemanticSource(t *testing.T) {
	ir := acceptanceFixture()
	ir.Activities[0].Source = SourceSpan{URI: "main.gooo", Start: Position{Line: 4, Column: 1}, End: Position{Line: 9, Column: 1}}
	ir.Activities[0].Slots[0].Source = SourceSpan{URI: "main.gooo", Start: Position{Line: 7, Column: 3}, End: Position{Line: 7, Column: 24}}
	first := mustAcceptanceResult(t, ir, nil)
	renamed := ir
	renamed.Activities[0].Name = "CompileBootstrap"
	renamed.Activities[0].GoName = "CompileBootstrap"
	second := mustAcceptanceResult(t, renamed, first.Source)
	assertStableMapping(t, first.SourceMap, second.SourceMap, "gooo://activity/compile")
	assertStableMapping(t, first.SourceMap, second.SourceMap, "gooo://slot/compile-implementation")
}

func TestFixtureImportPermutationIsReproducible(t *testing.T) {
	firstIR := acceptanceFixture()
	firstIR.Imports = []Import{{Name: "_", Path: "example/z"}, {Name: "_", Path: "example/a"}, {Name: "_", Path: "example/m"}}
	secondIR := firstIR
	secondIR.Imports = []Import{{Name: "_", Path: "example/m"}, {Name: "_", Path: "example/z"}, {Name: "_", Path: "example/a"}}
	first := mustAcceptanceResult(t, firstIR, nil)
	second := mustAcceptanceResult(t, secondIR, nil)
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.SourceMap, second.SourceMap) {
		t.Fatal("import permutation changed generated source or source map")
	}
}

func assertStableMapping(t *testing.T, first, second SourceMap, id string) {
	t.Helper()
	left, right := first.Lookup(id), second.Lookup(id)
	if len(left) != 1 || len(right) != 1 {
		t.Fatalf("stable ID %q has unexpected mappings: %#v %#v", id, left, right)
	}
	if left[0].SemanticID != right[0].SemanticID || left[0].Source != right[0].Source {
		t.Fatalf("stable ID %q changed source identity: %#v %#v", id, left[0], right[0])
	}
}
