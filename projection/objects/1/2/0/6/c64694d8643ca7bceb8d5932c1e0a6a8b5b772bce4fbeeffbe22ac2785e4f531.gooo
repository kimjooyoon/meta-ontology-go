package analyzer

// Candidate is a potentially semantic relation that could not be selected
// deterministically. Options are sorted by stable identity.
type Candidate struct {
	Subject   Identity
	Relation  Relation
	Reference string
	Options   []Identity
	Span      Span
	Reason    string
	Origin    ObservationOrigin
}

// IdentityState explains why an implementation observation stayed deferred.
// These states are never semantic fact statuses and cannot enter candidates.
type IdentityState string

const (
	IdentityUnresolved IdentityState = "unresolved"
	IdentityAmbiguous  IdentityState = "ambiguous"
	IdentityInvalid    IdentityState = "invalid"
)

func (s IdentityState) valid() bool {
	return s == IdentityUnresolved || s == IdentityAmbiguous || s == IdentityInvalid
}

// ImplementationDetail records a source observation that stayed in the Go
// view because it has no usable registered semantic identity.
type ImplementationDetail struct {
	Reference     string        `json:"reference"`
	Span          Span          `json:"span"`
	Reason        string        `json:"reason"`
	IdentityState IdentityState `json:"identity_state"`
}

func (d ImplementationDetail) normalized() ImplementationDetail {
	if d.IdentityState == "" {
		d.IdentityState = IdentityUnresolved
	}
	return d
}

// SemanticDelta is the output of one analysis. Added contains only
// source-backed deterministic facts; ambiguity and implementation details are
// kept in separate collections.
type SemanticDelta struct {
	Added                 []Fact
	Candidates            []Candidate
	ImplementationDetails []ImplementationDetail
}

// DeterministicFacts returns a copy of the facts in the delta.
func (d SemanticDelta) DeterministicFacts() []Fact {
	return append([]Fact(nil), d.Added...)
}

// Result contains the semantic delta and source-local registrations discovered
// from annotations.
type Result struct {
	Delta         SemanticDelta
	Registrations []Registration
	Diagnostics   Diagnostics
}
