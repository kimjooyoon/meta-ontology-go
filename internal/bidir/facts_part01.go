package bidir

import (
	"fmt"
)

// FactLayer records how strongly a fact is supported.
type FactLayer uint8

const (
	SyntacticFact FactLayer = iota + 1
	DeterministicFact
	CandidateFact
)

func (l FactLayer) String() string {
	switch l {
	case SyntacticFact:
		return "syntactic"
	case DeterministicFact:
		return "deterministic"
	case CandidateFact:
		return "candidate"
	default:
		return fmt.Sprintf("layer(%d)", l)
	}
}

// Fact is the adapter format for parser and Go analyzer observations.
type Fact struct {
	Layer     FactLayer
	Subject   ID
	Predicate Predicate
	Object    ID
	// EvidenceID is an adapter-supplied provenance identifier. It is kept out
	// of FactKey and semantic canonicalization so same-edge observations can
	// retain distinct evidence records without changing fact identity.
	EvidenceID  string
	SubjectKind Kind
	ObjectKind  Kind
	Attributes  map[string]string
	Source      SourceSpan
	Reason      string
}

// NewFact creates a fact without source metadata.
func NewFact(layer FactLayer, subject ID, predicate Predicate, object ID) Fact {
	return Fact{Layer: layer, Subject: subject, Predicate: predicate, Object: object}
}

// NewSourcedFact creates a source-backed fact.
func NewSourcedFact(layer FactLayer, subject ID, predicate Predicate, object ID, source SourceSpan) Fact {
	fact := NewFact(layer, subject, predicate, object)
	fact.Source = source
	return fact
}

// FactKey identifies a fact including its confidence layer.
type FactKey struct {
	Layer     FactLayer
	Subject   ID
	Predicate Predicate
	Object    ID
}

// Key returns layer-sensitive identity.
func (f Fact) Key() FactKey {
	return FactKey{Layer: f.Layer, Subject: f.Subject, Predicate: f.Predicate, Object: f.Object}
}

// SemanticKey returns layer-independent edge identity.
func (f Fact) SemanticKey() string {
	return relationKey(f.Predicate, f.Subject, f.Object)
}
