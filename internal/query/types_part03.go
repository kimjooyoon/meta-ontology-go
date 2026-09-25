package query

import (
	"fmt"
	"strings"
)

// Fact is one directed relation between stable semantic IDs. Reason is
// optional context for a candidate and never changes triple identity.
type Fact struct {
	Subject   ID         `json:"subject"`
	Predicate Relation   `json:"predicate"`
	Object    ID         `json:"object"`
	Status    FactStatus `json:"status"`
	Reason    string     `json:"reason,omitempty"`
}

func NewFact(subject ID, predicate Relation, object ID) Fact {
	return Fact{Subject: subject, Predicate: predicate, Object: object, Status: FactDeterministic}
}
func NewCandidateFact(subject ID, predicate Relation, object ID, reason string) Fact {
	return Fact{Subject: subject, Predicate: predicate, Object: object, Status: FactCandidate, Reason: reason}
}

// Key identifies a triple independently of fact status. A deterministic fact
// shadows a candidate with the same key.
type FactKey struct {
	Subject   ID
	Predicate Relation
	Object    ID
}

func (fact Fact) Key() FactKey {
	return FactKey{Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object}
}

// Normalized returns a validated, canonical copy of fact.
func (fact Fact) Normalized() (Fact, error) {
	subject, err := ParseID(fact.Subject.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: subject: %v", ErrInvalidFact, err)
	}
	object, err := ParseID(fact.Object.String())
	if err != nil {
		return Fact{}, fmt.Errorf("%w: object: %v", ErrInvalidFact, err)
	}
	predicate, err := ParseRelation(fact.Predicate)
	if err != nil {
		return Fact{}, fmt.Errorf("%w: predicate: %v", ErrInvalidFact, err)
	}
	if fact.Status == 0 {
		fact.Status = FactDeterministic
	}
	if fact.Status != FactDeterministic && fact.Status != FactCandidate {
		return Fact{}, fmt.Errorf("%w: unknown status %d", ErrInvalidFact, fact.Status)
	}
	fact.Subject = subject
	fact.Predicate = predicate
	fact.Object = object
	fact.Reason = strings.TrimSpace(fact.Reason)
	return fact, nil
}

// ExactQuery describes one complete triple. Empty values are invalid rather
// than acting as accidental wildcards; wildcard search is outside this small
// engine's contract.
type ExactQuery struct {
	Subject   ID
	Predicate Relation
	Object    ID
}
