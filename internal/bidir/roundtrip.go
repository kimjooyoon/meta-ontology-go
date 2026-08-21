package bidir

import "github.com/kimjooyoon/meta-ontology-go/internal/semantic"

// EquivalentAfterRoundTrip compares semantic meaning, not source formatting.
func EquivalentAfterRoundTrip(original, roundTripped semantic.IR) bool {
	return original.SemanticCanonical() == roundTripped.SemanticCanonical()
}

// PromoteCandidate returns a normalized IR with one reviewed candidate made
// deterministic. The input IR is not mutated when promotion fails or succeeds.
func PromoteCandidate(original semantic.IR, key semantic.FactKey) (semantic.IR, semantic.Fact, error) {
	ir, err := original.Normalized()
	if err != nil {
		return semantic.IR{}, semantic.Fact{}, err
	}
	promoted, err := ir.Graph.PromoteCandidate(key)
	if err != nil {
		return semantic.IR{}, semantic.Fact{}, err
	}
	if err := ir.Validate(); err != nil {
		return semantic.IR{}, semantic.Fact{}, err
	}
	return ir, promoted, nil
}
