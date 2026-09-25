package query

import (
	"fmt"
)

func NewExactQuery(subject ID, predicate Relation, object ID) ExactQuery {
	return ExactQuery{Subject: subject, Predicate: predicate, Object: object}
}
func (query ExactQuery) normalized() (ExactQuery, error) {
	fact, err := NewFact(query.Subject, query.Predicate, query.Object).Normalized()
	if err != nil {
		return ExactQuery{}, fmt.Errorf("%w: %v", ErrInvalidQuery, err)
	}
	return ExactQuery{Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object}, nil
}

// MatchResult keeps authoritative and review-needed matches separate.
type MatchResult struct {
	Deterministic []Fact
	Candidates    []Fact
	Metadata      ProjectionMetadata
}

func (result MatchResult) Empty() bool {
	return len(result.Deterministic) == 0 && len(result.Candidates) == 0
}
func (result MatchResult) All() []Fact {
	facts := append([]Fact(nil), result.Deterministic...)
	facts = append(facts, result.Candidates...)
	sortFacts(facts)
	return facts
}

// Direction controls which side of a directed relation traversal follows.
type Direction uint8

const (
	Outgoing Direction = iota + 1
	Incoming
	Both
)

// TraversalOptions bounds a traversal. An empty Predicate follows every PROV
// relation; zero Direction defaults to Outgoing; zero Selection includes both
// fact layers.
type TraversalOptions struct {
	Predicate Relation
	Direction Direction
	MaxDepth  int
	// Limit bounds both returned paths and edge inspection. Zero preserves
	// the unbounded direct Go API; query envelopes always set this field.
	Limit     int
	Selection FactSelection
}

// Path is a simple path beginning at the requested start ID. IDs are ordered
// in traversal direction and Facts contain the canonical relation direction.
type Path struct {
	IDs    []ID       `json:"ids"`
	Facts  []Fact     `json:"facts"`
	Status FactStatus `json:"status"`
}

func (path Path) Depth() int { return len(path.Facts) }
func (path Path) Last() ID {
	if len(path.IDs) == 0 {
		return ""
	}
	return path.IDs[len(path.IDs)-1]
}
