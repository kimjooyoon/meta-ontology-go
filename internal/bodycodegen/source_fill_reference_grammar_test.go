package bodycodegen

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func TestDerivedReferenceGrammarRequiresVersionedClosedProductions(t *testing.T) {
	grammars := []string{
		"integer-offset-constant/v1", "integer-predicate/v1", "integer-predicate-composition/v1",
		"integer-predicate-outside-range/v1", "integer-predicate-cutpoint/v1",
		"record-field-predicate/v1", "record-field-predicate/v2", "record-field-relation/v1",
		"record-field-relation-composition/v1", "record-field-relation-composition/v2",
		"record-field-relation-composition/v3", "record-field-predicate-composition/v1",
		"record-field-predicate-composition/v2", "record-string-literal/v1",
		"record-integer-literal/v1", "record-boolean-literal/v1",
	}
	for _, grammar := range grammars {
		if err := ValidateCallFreeFillGeneration(&assemblyspec.FillGeneration{Grammar: grammar}); err != nil {
			t.Fatal(grammar, err)
		}
	}
	for _, generation := range []*assemblyspec.FillGeneration{nil,
		{Grammar: "activity-call/v1"},
		{HoleGrammars: []assemblyspec.FillHoleGrammar{{HoleID: "condition", Grammar: "integer-predicate/v1"},
			{HoleID: "value", Grammar: "integer-offset-constant/v2"}}},
	} {
		if err := ValidateCallFreeFillGeneration(generation); err == nil || generation != nil &&
			!strings.Contains(err.Error(), "requires a candidate reference-scope map") {
			t.Fatal("unknown productions acquired reference authority", generation, err)
		}
	}
	t.Logf("REFERENCE_GRAMMAR_PROFILE_OBSERVATION known_versions=%d rejected_profiles=%d enumeration=0 scoring=0 model_predictions=0 program_execution=0", len(grammars), 3)
}
