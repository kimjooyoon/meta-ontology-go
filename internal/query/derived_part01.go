package query

import (
	"errors"
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
