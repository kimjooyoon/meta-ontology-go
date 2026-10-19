package syntaxregistration

import (
	"context"
	"testing"
)

func TestProfileExerciseDoesNotTurnUnresolvedAliasIntoAbsence(t *testing.T) {
	source := []byte(`package p namespace n
entity Text id "urn:gooo:type:string"
entity E id "urn:e" fields {
 field F id "urn:f" type Text required many
}`)
	unknown := observeProfileExercise("alias.gooo", source, "entity-string")
	if unknown.Known || unknown.MatchCount != 0 || unknown.FirstWitness != nil ||
		unknown.Reason != "TYPE_REFERENCE_UNRESOLVED" ||
		expectedExerciseStatus("entity-string", unknown) != "UNKNOWN" {
		t.Fatalf("unresolved type alias was treated as absent: %+v", unknown)
	}
	many := observeProfileExercise("alias.gooo", source, "entity-many")
	if !many.Known || many.MatchCount != 1 || many.FirstWitness == nil ||
		expectedExerciseStatus("entity-many", many) != "EXERCISED" {
		t.Fatalf("unrelated type alias hid a directly authored many field: %+v", many)
	}
	direct := []byte(`package p namespace n
entity E id "urn:e" fields {
 field S id "urn:s" type string required one
 field A id "urn:a" type Alias required one
}`)
	positive := observeProfileExercise("mixed.gooo", direct, "entity-string")
	if !positive.Known || positive.MatchCount != 1 || positive.FirstWitness == nil ||
		expectedExerciseStatus("entity-string", positive) != "EXERCISED" {
		t.Fatalf("direct source witness was hidden by another unresolved spelling: %+v", positive)
	}
}

func TestProfileExerciseExplainsUnreadableAndUnparseableSource(t *testing.T) {
	missing := InspectProfileHintForCapability(context.Background(), osProjectFS(),
		"missing.gooo", "entity-many")
	if missing.Exercise.Status != "UNKNOWN" || missing.Exercise.Reason != "SOURCE_UNREADABLE" {
		t.Fatalf("unreadable source lost its reason: %+v", missing.Exercise)
	}
	invalid := observeProfileExercise("invalid.gooo", []byte("package broken\nentity"), "entity-many")
	if invalid.Known || invalid.Reason != "SOURCE_PARSE_UNAVAILABLE" ||
		expectedExerciseStatus("entity-many", invalid) != "UNKNOWN" {
		t.Fatalf("unparseable source became a negative observation: %+v", invalid)
	}
}
