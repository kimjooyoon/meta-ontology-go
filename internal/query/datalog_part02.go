package query

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
