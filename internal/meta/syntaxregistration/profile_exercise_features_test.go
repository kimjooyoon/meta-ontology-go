package syntaxregistration

import (
	"strings"
	"testing"
)

func TestProfileExerciseObservesEachAuthoredEntityCapability(t *testing.T) {
	source := []byte(`package p namespace n
entity E id "urn:e" fields {
 field S id "urn:s" type string required one max_utf8_bytes "8"
 field B id "urn:b" type boolean optional one
 field I id "urn:i" type integer required many
}`)
	for _, row := range []struct {
		requirement, fieldID string
	}{
		{"entity-string", "urn:s"},
		{"entity-boolean", "urn:b"},
		{"entity-integer", "urn:i"},
		{"entity-optional", "urn:b"},
		{"entity-many", "urn:i"},
		{"entity-text-bound", "urn:s"},
	} {
		t.Run(row.requirement, func(t *testing.T) {
			observed := observeProfileExercise("all.gooo", source, row.requirement)
			witness := observed.FirstWitness
			if !observed.Known || observed.MatchCount != 1 || witness == nil ||
				expectedExerciseStatus(row.requirement, observed) != "EXERCISED" ||
				witness.EntityID != "urn:e" || witness.FieldID != row.fieldID ||
				witness.StartByte < 0 || witness.EndByte > len(source) ||
				witness.StartByte >= witness.EndByte ||
				!strings.Contains(string(source[witness.StartByte:witness.EndByte]), row.fieldID) {
				t.Fatalf("source field witness mismatch: %+v", observed)
			}
		})
	}
}
