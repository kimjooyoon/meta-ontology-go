package query

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

var (
	ErrInvalidID        = errors.New("invalid stable ID")
	ErrInvalidRelation  = errors.New("invalid PROV relation")
	ErrInvalidFact      = errors.New("invalid query fact")
	ErrInvalidQuery     = errors.New("invalid exact query")
	ErrInvalidTraversal = errors.New("invalid traversal options")
)

// StableID is a canonical, URI-like semantic identity. Display names are not
// accepted as substitutes because equal names may belong to different scopes.
type StableID string

// ID is the shorter spelling used by facts and query APIs.
type ID = StableID

// ParseID validates and canonicalizes a stable ID. URI scheme and host casing
// are normalized; path and opaque URI content retain their case.
func ParseID(raw string) (StableID, error) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", fmt.Errorf("%w: empty or whitespace-containing value", ErrInvalidID)
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidID, err)
	}
	if parsed.Scheme == "" {
		return "", fmt.Errorf("%w: URI scheme is required", ErrInvalidID)
	}
	if strings.Contains(value, "://") && parsed.Host == "" {
		return "", fmt.Errorf("%w: URI authority is empty", ErrInvalidID)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	canonical := parsed.String()
	if canonical == "" {
		return "", fmt.Errorf("%w: empty canonical value", ErrInvalidID)
	}
	return StableID(canonical), nil
}

// NewID is an alias for ParseID for callers constructing graph inputs.
func NewID(raw string) (ID, error) { return ParseID(raw) }

func (id StableID) String() string { return string(id) }

// Valid reports whether id is already a valid stable identity. It does not
// require callers to retain the canonical form when only checking input.
func (id StableID) Valid() bool {
	_, err := ParseID(id.String())
	return err == nil
}

// Relation is a directed PROV relation. Its subject and object follow the
// PROV-O direction, for example Activity prov:used Entity.
type Relation string

const (
	Used              Relation = "used"
	WasGeneratedBy    Relation = "wasGeneratedBy"
	WasDerivedFrom    Relation = "wasDerivedFrom"
	WasAssociatedWith Relation = "wasAssociatedWith"

	RelationUsed              = Used
	RelationWasGeneratedBy    = WasGeneratedBy
	RelationWasDerivedFrom    = WasDerivedFrom
	RelationWasAssociatedWith = WasAssociatedWith
)

// PROV-prefixed spellings are accepted at serialization/query boundaries and
// normalize to the local names used by the semantic IR.
const (
	PROVUsed              Relation = "prov:used"
	PROVWasGeneratedBy    Relation = "prov:wasGeneratedBy"
	PROVWasDerivedFrom    Relation = "prov:wasDerivedFrom"
	PROVWasAssociatedWith Relation = "prov:wasAssociatedWith"
)

// ParseRelation accepts canonical PROV names and their compact spellings.
func ParseRelation(raw Relation) (Relation, error) {
	switch strings.TrimSpace(string(raw)) {
	case string(Used), string(PROVUsed):
		return Used, nil
	case string(WasGeneratedBy), string(PROVWasGeneratedBy):
		return WasGeneratedBy, nil
	case string(WasDerivedFrom), string(PROVWasDerivedFrom):
		return WasDerivedFrom, nil
	case string(WasAssociatedWith), string(PROVWasAssociatedWith):
		return WasAssociatedWith, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidRelation, raw)
	}
}

func (relation Relation) String() string { return string(relation) }

func (relation Relation) Valid() bool {
	_, err := ParseRelation(relation)
	return err == nil
}

// FactStatus separates authoritative facts from observations that still need
// an assertion or review.
type FactStatus uint8

const (
	FactDeterministic FactStatus = iota + 1
	FactCandidate

	Deterministic = FactDeterministic
	Candidate     = FactCandidate
)

func (status FactStatus) String() string {
	switch status {
	case FactDeterministic:
		return "deterministic"
	case FactCandidate:
		return "candidate"
	default:
		return "unknown"
	}
}

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

// TraversalResult keeps paths containing candidate facts separate from paths
// made entirely of deterministic facts.
type TraversalResult struct {
	Deterministic []Path
	Candidates    []Path
	Metadata      ProjectionMetadata
}

func (result TraversalResult) All() []Path {
	paths := append([]Path(nil), result.Deterministic...)
	paths = append(paths, result.Candidates...)
	sortPaths(paths)
	return paths
}
