package query

import (
	"errors"
)

const (
	DefaultDatalogLimit        = 100
	MaxDatalogLimit            = MaxEnvelopeLimit
	DefaultDatalogDerivedLimit = 10000
	DefaultDatalogDepth        = MaxEnvelopeDepth
	MaxDatalogDepth            = MaxEnvelopeDepth
	DefaultDatalogWork         = 10000
	MaxDatalogWork             = 100000
	MaxDatalogRules            = 128
	MaxDatalogBodyAtoms        = 32
)

var (
	ErrInvalidDatalogQuery = errors.New("invalid Datalog query")
	ErrDatalogBudget       = errors.New("Datalog evaluation budget exceeded")
)

// DatalogTerm is either a stable ID constant or a variable. Variables are
// canonicalized with a leading '?' but callers may pass either "name" or
// "?name" to Variable.
type DatalogTerm struct {
	Variable string
	Constant ID
}

// Term is the concise spelling used by parser-neutral callers.
type Term = DatalogTerm

func Variable(name string) DatalogTerm { return DatalogTerm{Variable: name} }
func Constant(id ID) DatalogTerm       { return DatalogTerm{Constant: id} }

// DatalogAtom is a positive binary triple pattern. Negation, aggregation, and
// unbounded path operators are intentionally not part of this query boundary.
type DatalogAtom struct {
	Predicate string
	Subject   DatalogTerm
	Object    DatalogTerm
}
type Atom = DatalogAtom

func Triple(predicate string, subject, object DatalogTerm) DatalogAtom {
	return DatalogAtom{Predicate: predicate, Subject: subject, Object: object}
}

// DatalogRule is a positive implication. Every variable in Head must occur in
// Body, which makes rule evaluation finite over the selected fact universe.
type DatalogRule struct {
	ID   string
	Head DatalogAtom
	Body []DatalogAtom
}
type Rule = DatalogRule
