package query

import (
	"errors"
	"fmt"
)

const DerivedRuleSchemaVersion = "gooo-query/rules/v1"

const (
	DerivedStatusNotRequested     = "not_requested"
	DerivedStatusNonAuthoritative = "derived_non_authoritative"
	DerivedFactStatus             = DerivedStatusNonAuthoritative
)

var (
	ErrUnsupportedDerivedRule = errors.New("unsupported derived query rule")
	ErrInvalidDerivedQuery    = errors.New("invalid derived query")
)

// DerivedRuleID is a stable, versioned rule identity. Inverse rules map used
// to usedBy, wasGeneratedBy to generatedBy, and wasDerivedFrom to derivedTo;
// dependsOn follows wasDerivedFrom transitively. Rule IDs never become facts.
type DerivedRuleID string

const (
	RuleUsedBy      DerivedRuleID = DerivedRuleSchemaVersion + "/inverse/usedBy"
	RuleGeneratedBy DerivedRuleID = DerivedRuleSchemaVersion + "/inverse/generatedBy"
	RuleDerivedTo   DerivedRuleID = DerivedRuleSchemaVersion + "/inverse/derivedTo"
	RuleDependsOn   DerivedRuleID = DerivedRuleSchemaVersion + "/transitive/dependsOn"
)

// DerivedRelation names a non-authoritative view relation.
type DerivedRelation string

const (
	DerivedUsedBy      DerivedRelation = "usedBy"
	DerivedGeneratedBy DerivedRelation = "generatedBy"
	DerivedTo          DerivedRelation = "derivedTo"
	DerivedDependsOn   DerivedRelation = "dependsOn"
)

// DerivedFact is a read-only rule result. Status is always explicit and
// non-authoritative; SourceLayer records whether the supporting graph facts
// were deterministic or candidate facts.
type DerivedFact struct {
	Subject     ID              `json:"subject"`
	Predicate   DerivedRelation `json:"predicate"`
	Object      ID              `json:"object"`
	RuleID      DerivedRuleID   `json:"rule_id"`
	Depth       int             `json:"depth"`
	Status      string          `json:"status"`
	SourceLayer string          `json:"source_layer"`
}

// DerivedOptions bounds a rule evaluation over a selected graph layer.
type DerivedOptions struct {
	Rule      DerivedRuleID
	MaxDepth  int
	Limit     int
	Selection FactSelection
}

// DerivedResult keeps derived rows separate from canonical and candidate
// facts. It is a detached result and does not change Graph or GraphHash.
type DerivedResult struct {
	Deterministic []DerivedFact
	Candidates    []DerivedFact
	Metadata      ProjectionMetadata
}

type derivedRule struct {
	id         DerivedRuleID
	predicate  DerivedRelation
	base       Relation
	inverse    bool
	transitive bool
}

type derivedKey struct {
	subject   ID
	predicate DerivedRelation
	object    ID
}

func ParseDerivedRule(raw DerivedRuleID) (DerivedRuleID, error) {
	switch raw {
	case RuleUsedBy, RuleGeneratedBy, RuleDerivedTo, RuleDependsOn:
		return raw, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedDerivedRule, raw)
	}
}

func ruleDefinition(ruleID DerivedRuleID) (derivedRule, error) {
	switch ruleID {
	case RuleUsedBy:
		return derivedRule{ruleID, DerivedUsedBy, Used, true, false}, nil
	case RuleGeneratedBy:
		return derivedRule{ruleID, DerivedGeneratedBy, WasGeneratedBy, true, false}, nil
	case RuleDerivedTo:
		return derivedRule{ruleID, DerivedTo, WasDerivedFrom, true, false}, nil
	case RuleDependsOn:
		return derivedRule{ruleID, DerivedDependsOn, WasDerivedFrom, false, true}, nil
	default:
		return derivedRule{}, fmt.Errorf("%w: %q", ErrUnsupportedDerivedRule, ruleID)
	}
}

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
