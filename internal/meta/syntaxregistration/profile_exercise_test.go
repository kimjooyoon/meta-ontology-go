package syntaxregistration

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestProfileExerciseSeparatesCompatibilityFromAuthoredFields(t *testing.T) {
	repository := osProjectFS()
	plain := InspectProfileHintForCapability(context.Background(), repository,
		"examples/billing/main.gooo", "entity-many")
	if plain.Status != "MINIMAL_HINT" || plain.Exercise.Status != "NOT_EXERCISED" ||
		!plain.Exercise.Known || plain.Exercise.MatchCount != 0 || plain.Exercise.FirstWitness != nil {
		t.Fatalf("compatible profile was confused with exercised source feature: %+v", plain)
	}
	path := "examples/collection-filter/main.gooo"
	observed := InspectProfileHintForCapability(context.Background(), repository, path, "entity-many")
	witness := observed.Exercise.FirstWitness
	raw, err := os.ReadFile("../../../" + path)
	if err != nil {
		t.Fatal(err)
	}
	if observed.Exercise.Status != "EXERCISED" || observed.Exercise.MatchCount != 1 ||
		witness == nil || witness.EntityID != "urn:collection-filter:labels" ||
		witness.FieldID != "urn:collection-filter:labels:values" ||
		witness.StartByte < 0 || witness.EndByte > len(raw) || witness.StartByte >= witness.EndByte ||
		!strings.Contains(string(raw[witness.StartByte:witness.EndByte]), "required many") ||
		observed.SourceDigest != digest(raw) {
		t.Fatalf("source feature lost its pinned field witness: %+v", observed)
	}
	withoutRequest := InspectProfileHint(context.Background(), repository, path)
	if withoutRequest.Exercise.Status != "NOT_REQUESTED" {
		t.Fatalf("unsolicited source exercise: %+v", withoutRequest.Exercise)
	}
	route := InspectProfileHintForCapability(context.Background(), repository, path, "ordered-collections")
	if route.Exercise.Status != "ROUTE_ONLY" {
		t.Fatalf("parser route was treated as one field feature: %+v", route.Exercise)
	}
}

func TestProfileExerciseGoooStatusIsCheckedAgainstPinnedSource(t *testing.T) {
	path := "examples/collection-filter/main.gooo"
	changed := strings.Replace(string(contractSource),
		`if input.exercise_count > 0 { exercise_status = "EXERCISED" }`,
		`if input.exercise_count > 0 { exercise_status = "NOT_EXERCISED" }`, 1)
	if changed == string(contractSource) {
		t.Fatal("Gooo exercise decision was not found")
	}
	report := inspectProfileHintSourceForCapability(context.Background(), osProjectFS(),
		path, "entity-many", []byte(changed))
	if report.Status != "UNKNOWN" || report.Reason != "HINT_CONTRACT_MISMATCH" ||
		report.Exercise.Status != "UNKNOWN" {
		t.Fatalf("unchecked Gooo exercise decision was trusted: %+v", report)
	}
}

func TestProfileExerciseUnknownSyntaxAndStableTypeID(t *testing.T) {
	unknown := observeProfileExercise("bad.gooo", []byte("package broken\nentity"), "entity-many")
	if unknown.Known || expectedExerciseStatus("entity-many", unknown) != "UNKNOWN" {
		t.Fatalf("invalid source became a negative observation: %+v", unknown)
	}
	source := []byte(`package p namespace n entity E id "urn:e" fields {
 field F id "urn:f" type "urn:gooo:type:string" required one
}`)
	observed := observeProfileExercise("stable.gooo", source, "entity-string")
	if !observed.Known || observed.MatchCount != 1 || observed.FirstWitness == nil ||
		expectedExerciseStatus("entity-string", observed) != "EXERCISED" {
		t.Fatalf("stable builtin type identity was missed: %+v", observed)
	}
}
