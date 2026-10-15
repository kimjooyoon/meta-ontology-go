package bodycodegen

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// ValidateCallFreeFillGeneration identifies versioned expression grammars whose
// productions contain only literals, input selectors and value operators. They
// create neither calls nor lexical bindings. This is a reference-scope property,
// not candidate enumeration, type checking, selection or functional evaluation.
func ValidateCallFreeFillGeneration(generation *assemblyspec.FillGeneration) error {
	if generation == nil {
		return fmt.Errorf("source_fill generation is absent")
	}
	if len(generation.HoleGrammars) == 0 {
		return validateCallFreeFillGrammar(generation.Grammar)
	}
	for _, hole := range generation.HoleGrammars {
		if err := validateCallFreeFillGrammar(hole.Grammar); err != nil {
			return fmt.Errorf("source_fill hole %q: %w", hole.HoleID, err)
		}
	}
	return nil
}

func validateCallFreeFillGrammar(grammar string) error {
	switch grammar {
	case bodySearchIntegerAffineGrammar,
		"integer-predicate/v1", "integer-predicate-composition/v1",
		"integer-predicate-outside-range/v1", "integer-predicate-cutpoint/v1",
		recordFieldPredicateGrammar, recordFieldPredicateV2Grammar,
		recordFieldRelationGrammar, recordFieldRelationCompositionGrammar,
		recordFieldRelationCompositionV2Grammar, recordFieldRelationCompositionV3Grammar,
		recordPredicateCompositionGrammar, recordPredicateCompositionV2Grammar,
		recordStringLiteralGrammar, recordIntegerLiteralGrammar, recordBooleanLiteralGrammar:
		return nil
	default:
		return fmt.Errorf("source_fill grammar %q requires a candidate reference-scope map", grammar)
	}
}
