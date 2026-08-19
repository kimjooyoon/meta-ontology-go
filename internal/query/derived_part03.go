package query

import (
	"fmt"
)

// Derive evaluates one registered rule from a stable root. It never inserts
// derived rows into the graph and never changes the graph fingerprint.
func (graph Graph) Derive(root ID, options DerivedOptions) (DerivedResult, error) {
	canonicalRoot, err := ParseID(root.String())
	if err != nil {
		return DerivedResult{}, err
	}
	normalized, rule, err := normalizeDerivedOptions(options)
	if err != nil {
		return DerivedResult{}, err
	}
	if err := graph.requireEndpoint(canonicalRoot); err != nil {
		return DerivedResult{}, err
	}
	var deterministic, candidates []DerivedFact
	if rule.transitive {
		deterministic, candidates = graph.deriveDependsOn(canonicalRoot, normalized)
	} else {
		deterministic, candidates = graph.deriveInverse(canonicalRoot, rule, normalized)
	}
	deterministic, candidates = limitDerived(deterministic, candidates, normalized.Limit)
	metadata := graph.Metadata()
	metadata.DerivedStatus = DerivedStatusNonAuthoritative
	metadata.DerivedRuleSchema = DerivedRuleSchemaVersion
	metadata.DerivedRuleDigest, err = normalized.Rule.CanonicalDigest()
	if err != nil {
		return DerivedResult{}, err
	}
	for index := range metadata.AuthorityLabels {
		if metadata.AuthorityLabels[index].View == "derived_query" {
			metadata.AuthorityLabels[index].Status = DerivedStatusNonAuthoritative
		}
	}
	return DerivedResult{Deterministic: deterministic, Candidates: candidates, Metadata: metadata}, nil
}
func normalizeDerivedOptions(options DerivedOptions) (DerivedOptions, derivedRule, error) {
	ruleID, err := ParseDerivedRule(options.Rule)
	if err != nil {
		return DerivedOptions{}, derivedRule{}, err
	}
	rule, err := ruleDefinition(ruleID)
	if err != nil {
		return DerivedOptions{}, derivedRule{}, err
	}
	if options.MaxDepth < 1 || options.MaxDepth > MaxEnvelopeDepth {
		return DerivedOptions{}, derivedRule{}, fmt.Errorf(
			"%w: max depth must be 1..%d", ErrInvalidDerivedQuery, MaxEnvelopeDepth,
		)
	}
	if options.Limit < 1 || options.Limit > MaxEnvelopeLimit {
		return DerivedOptions{}, derivedRule{}, fmt.Errorf(
			"%w: row limit must be 1..%d", ErrInvalidDerivedQuery, MaxEnvelopeLimit,
		)
	}
	if rule.inverse && options.MaxDepth != 1 {
		return DerivedOptions{}, derivedRule{}, fmt.Errorf(
			"%w: inverse rule depth must be 1", ErrInvalidDerivedQuery,
		)
	}
	selection, err := options.Selection.normalized()
	if err != nil {
		return DerivedOptions{}, derivedRule{}, err
	}
	options.Rule = ruleID
	options.Selection = selection
	return options, rule, nil
}
