package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

func Constant(id ID) DatalogTerm { return DatalogTerm{Constant: id} }

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

// DatalogQuery evaluates a conjunction of patterns. Candidates are visible to
// matching only when IncludeCandidates is true; they are never rule inputs.
type DatalogQuery struct {
	Patterns          []DatalogAtom
	Rules             []DatalogRule
	IncludeCandidates bool
	IncludeDerived    bool
	Limit             int
	MaxDerivedFacts   int
	// MaxDepth bounds recursive rule derivation depth. Zero selects the
	// versioned safe default; it never means unbounded evaluation.
	MaxDepth int
	// MaxWork bounds rule-body and pattern candidate inspections. Zero selects
	// the versioned safe default; it never means unbounded evaluation.
	MaxWork int
}

// DatalogOrigin identifies the authority boundary of a returned fact.
type DatalogOrigin uint8

const (
	DatalogDeclared DatalogOrigin = iota + 1
	DatalogCandidate
	DatalogDerived
)

func (origin DatalogOrigin) String() string {
	switch origin {
	case DatalogDeclared:
		return "declared"
	case DatalogCandidate:
		return "candidate"
	case DatalogDerived:
		return "derived"
	default:
		return "unknown"
	}
}

// DatalogFact is a read-only fact in the selected query universe. Derived
// facts carry one deterministic support proof; they never enter Graph.
type DatalogFact struct {
	Namespace string
	Subject   ID
	Predicate string
	Object    ID
	Origin    DatalogOrigin
	RuleID    string
	Depth     int
	Support   []DatalogFactKey
}

type DatalogFactKey struct {
	Namespace string
	Subject   ID
	Predicate string
	Object    ID
}

func (fact DatalogFact) Key() DatalogFactKey {
	return DatalogFactKey{Namespace: fact.Namespace, Subject: fact.Subject, Predicate: fact.Predicate, Object: fact.Object}
}

// DatalogRow contains one set-semantics binding and the facts that satisfied
// the query patterns. Binding names omit the canonical '?' prefix.
type DatalogRow struct {
	Bindings map[string]ID
	Facts    []DatalogFact
}

func (row DatalogRow) Value(name string) (ID, bool) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "?")
	value, ok := row.Bindings[name]
	return value, ok
}

// DatalogResult is deterministic by construction. Complete is false when the
// result limit trimmed rows; no partial result is ever reported as complete.
type DatalogResult struct {
	Rows     []DatalogRow
	Derived  []DatalogFact
	Complete bool
}

// EvaluateDatalog evaluates positive rules over deterministic graph facts and
// then matches the requested patterns. It is a read-only projection and does
// not promote candidates or alter the graph hash.
func (graph Graph) EvaluateDatalog(request DatalogQuery) (DatalogResult, error) {
	normalized, rules, err := normalizeDatalogQuery(request)
	if err != nil {
		return DatalogResult{}, err
	}

	declared := make([]DatalogFact, 0, len(graph.DeterministicFacts()))
	for _, fact := range graph.DeterministicFacts() {
		declared = append(declared, DatalogFact{
			Namespace: graphScope(graph),
			Subject:   fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
			Origin: DatalogDeclared, Depth: 0,
		})
	}
	sortDatalogFacts(declared)
	var derived []DatalogFact
	workBudget := newDatalogWorkBudget(normalized.MaxWork)
	if normalized.IncludeDerived {
		derived, err = deriveDatalog(
			declared, rules, normalized.MaxDerivedFacts,
			normalized.MaxDepth, workBudget,
		)
		if err != nil {
			return DatalogResult{Complete: false}, err
		}
	}

	universe := append([]DatalogFact(nil), declared...)
	if normalized.IncludeDerived {
		universe = append(universe, derived...)
	}
	if normalized.IncludeCandidates {
		for _, fact := range graph.CandidateFacts() {
			universe = append(universe, DatalogFact{
				Namespace: graphScope(graph),
				Subject:   fact.Subject, Predicate: string(fact.Predicate), Object: fact.Object,
				Origin: DatalogCandidate, Depth: 0,
			})
		}
	}
	sortDatalogFacts(universe)
	rows, err := matchDatalogPatterns(normalized.Patterns, universe, workBudget)
	if err != nil {
		return DatalogResult{Complete: false}, err
	}
	complete := true
	if len(rows) > normalized.Limit {
		rows = rows[:normalized.Limit]
		complete = false
		return DatalogResult{Rows: rows, Derived: derived, Complete: false},
			fmt.Errorf("%w: maximum result rows %d", ErrDatalogBudget, normalized.Limit)
	}
	return DatalogResult{Rows: rows, Derived: derived, Complete: complete}, nil
}

func graphScope(graph Graph) string {
	if graph.binding == nil {
		return ""
	}
	return graph.binding.namespace
}

// CanonicalJSON is a stable request representation after defaults and all
// bounds have been normalized. It is suitable for replay/cache keys.
func (request DatalogQuery) CanonicalJSON() ([]byte, error) {
	normalized, _, err := normalizeDatalogQuery(request)
	if err != nil {
		return nil, err
	}
	return json.Marshal(normalized)
}

// CanonicalDigest hashes the normalized Datalog request bytes.
func (request DatalogQuery) CanonicalDigest() (string, error) {
	canonical, err := request.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

// CanonicalJSON returns the deterministic result receipt. It is not a source
// or authority hash; the graph hash remains owned by the semantic snapshot.
func (result DatalogResult) CanonicalJSON() ([]byte, error) {
	return json.Marshal(result)
}

// CanonicalDigest hashes the stable Datalog result bytes.
func (result DatalogResult) CanonicalDigest() (string, error) {
	canonical, err := result.CanonicalJSON()
	if err != nil {
		return "", err
	}
	return digestBytes(canonical), nil
}

// QueryDatalog is an API synonym that reads naturally at call sites.
func (graph Graph) QueryDatalog(request DatalogQuery) (DatalogResult, error) {
	return graph.EvaluateDatalog(request)
}
