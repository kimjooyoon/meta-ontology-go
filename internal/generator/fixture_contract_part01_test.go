package generator

import (
	"bytes"
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
