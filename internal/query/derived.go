package query

import (
	"errors"
	"fmt"
	"sort"
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

type derivedPath struct {
	node   ID
	ids    []ID
	status FactStatus
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
		deterministic, candidates = graph.deriveInverse(canonicalRoot, rule, normalized.Selection)
	}
	deterministic, candidates = limitDerived(deterministic, candidates, normalized.Limit)
	metadata := graph.Metadata()
	metadata.DerivedStatus = DerivedStatusNonAuthoritative
	metadata.DerivedRuleSchema = DerivedRuleSchemaVersion
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

func (graph Graph) deriveInverse(root ID, rule derivedRule, selection FactSelection) ([]DerivedFact, []DerivedFact) {
	deterministic := make(map[derivedKey]DerivedFact)
	candidates := make(map[derivedKey]DerivedFact)
	for _, fact := range graph.AllFacts() {
		if fact.Predicate != rule.base || fact.Object != root || !selection.includes(fact.Status) {
			continue
		}
		derived := newDerivedFact(rule, fact.Object, fact.Subject, 1, fact.Status)
		recordDerived(deterministic, candidates, derived)
	}
	return sortedDerived(deterministic), sortedDerived(candidates)
}

func (graph Graph) deriveDependsOn(root ID, options DerivedOptions) ([]DerivedFact, []DerivedFact) {
	deterministic := make(map[derivedKey]DerivedFact)
	candidates := make(map[derivedKey]DerivedFact)
	frontier := []derivedPath{{node: root, ids: []ID{root}, status: FactDeterministic}}
	for depth := 1; depth <= options.MaxDepth && len(frontier) > 0; depth++ {
		next := make([]derivedPath, 0)
		for _, path := range frontier {
			for _, fact := range graph.edges(path.node, TraversalOptions{
				Predicate: WasDerivedFrom, Direction: Outgoing, Selection: options.Selection,
			}, options.Selection) {
				if containsID(path.ids, fact.Object) {
					continue
				}
				status := path.status
				if fact.Status == FactCandidate {
					status = FactCandidate
				}
				derived := newDerivedFact(
					derivedRule{id: RuleDependsOn, predicate: DerivedDependsOn},
					root, fact.Object, depth, status,
				)
				recordDerived(deterministic, candidates, derived)
				next = append(next, derivedPath{
					node: fact.Object, ids: appendPathID(path.ids, fact.Object), status: status,
				})
			}
		}
		sortDerivedPaths(next)
		frontier = next
	}
	return sortedDerived(deterministic), sortedDerived(candidates)
}

func newDerivedFact(rule derivedRule, subject, object ID, depth int, status FactStatus) DerivedFact {
	return DerivedFact{
		Subject: subject, Predicate: rule.predicate, Object: object,
		RuleID: rule.id, Depth: depth, Status: DerivedFactStatus,
		SourceLayer: status.String(),
	}
}

func recordDerived(deterministic, candidates map[derivedKey]DerivedFact, fact DerivedFact) {
	key := derivedKey{fact.Subject, fact.Predicate, fact.Object}
	if fact.SourceLayer == FactCandidate.String() {
		if _, exists := deterministic[key]; !exists {
			candidates[key] = preferDerived(candidates[key], fact)
		}
		return
	}
	deterministic[key] = preferDerived(deterministic[key], fact)
	delete(candidates, key)
}

func preferDerived(existing, incoming DerivedFact) DerivedFact {
	if existing.RuleID == "" || incoming.Depth < existing.Depth {
		return incoming
	}
	return existing
}

func sortedDerived(rows map[derivedKey]DerivedFact) []DerivedFact {
	result := make([]DerivedFact, 0, len(rows))
	for _, row := range rows {
		result = append(result, row)
	}
	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
		if left.Subject != right.Subject {
			return left.Subject < right.Subject
		}
		if left.Predicate != right.Predicate {
			return left.Predicate < right.Predicate
		}
		if left.Object != right.Object {
			return left.Object < right.Object
		}
		if left.Depth != right.Depth {
			return left.Depth < right.Depth
		}
		return left.SourceLayer < right.SourceLayer
	})
	return result
}

func sortDerivedPaths(paths []derivedPath) {
	sort.Slice(paths, func(i, j int) bool {
		if len(paths[i].ids) != len(paths[j].ids) {
			return len(paths[i].ids) < len(paths[j].ids)
		}
		for index := range paths[i].ids {
			if paths[i].ids[index] != paths[j].ids[index] {
				return paths[i].ids[index] < paths[j].ids[index]
			}
		}
		return paths[i].status < paths[j].status
	})
}

func appendPathID(ids []ID, next ID) []ID {
	result := append([]ID(nil), ids...)
	return append(result, next)
}
