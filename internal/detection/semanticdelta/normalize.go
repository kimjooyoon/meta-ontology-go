package semanticdelta

import (
	"fmt"
	"sort"
	"strings"
)

// Normalized returns a validated, sorted, detached request.
func (r Request) Normalized() (Request, error) {
	version := strings.TrimSpace(r.Version)
	if version == "" {
		version = FormatVersion
	}
	if version != FormatVersion {
		return Request{}, fmt.Errorf("unsupported semanticdelta version %q", version)
	}
	scope, err := r.Allowed.Normalized()
	if err != nil {
		return Request{}, err
	}
	delta, err := r.Delta.Normalized()
	if err != nil {
		return Request{}, err
	}
	return Request{Version: version, Allowed: scope, Delta: delta}, nil
}

// Normalize validates and canonicalizes the request in place.
func (r *Request) Normalize() error {
	if r == nil {
		return fmt.Errorf("cannot normalize a nil request")
	}
	normalized, err := r.Normalized()
	if err != nil {
		return err
	}
	*r = normalized
	return nil
}

// Normalized returns a validated, sorted, detached scope.
func (s Scope) Normalized() (Scope, error) {
	ids, err := normalizeValues("scope ID", s.IDs)
	if err != nil {
		return Scope{}, err
	}
	prefixes, err := normalizeValues("scope prefix", s.Prefixes)
	if err != nil {
		return Scope{}, err
	}
	predicates, err := normalizeValues("scope predicate", s.Predicates)
	if err != nil {
		return Scope{}, err
	}
	return Scope{IDs: ids, Prefixes: prefixes, Predicates: predicates}, nil
}

// Normalize validates and canonicalizes the scope in place.
func (s *Scope) Normalize() error {
	if s == nil {
		return fmt.Errorf("cannot normalize a nil scope")
	}
	normalized, err := s.Normalized()
	if err != nil {
		return err
	}
	*s = normalized
	return nil
}

// Normalized returns a validated, sorted, detached snapshot.
func (s Snapshot) Normalized() (Snapshot, error) {
	nodes, err := normalizeNodes(s.Nodes, "snapshot nodes")
	if err != nil {
		return Snapshot{}, err
	}
	facts, err := normalizeFacts(s.Facts, "snapshot facts")
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{Nodes: nodes, Facts: facts}, nil
}

// Normalize validates and canonicalizes the snapshot in place.
func (s *Snapshot) Normalize() error {
	if s == nil {
		return fmt.Errorf("cannot normalize a nil snapshot")
	}
	normalized, err := s.Normalized()
	if err != nil {
		return err
	}
	*s = normalized
	return nil
}

// Normalized returns a validated, sorted, detached delta.
func (d Delta) Normalized() (Delta, error) {
	addedNodes, err := normalizeNodes(d.AddedNodes, "added nodes")
	if err != nil {
		return Delta{}, err
	}
	removedNodes, err := normalizeNodes(d.RemovedNodes, "removed nodes")
	if err != nil {
		return Delta{}, err
	}
	addedFacts, err := normalizeFacts(d.AddedFacts, "added facts")
	if err != nil {
		return Delta{}, err
	}
	removedFacts, err := normalizeFacts(d.RemovedFacts, "removed facts")
	if err != nil {
		return Delta{}, err
	}
	if overlapNodes(addedNodes, removedNodes) {
		return Delta{}, fmt.Errorf("delta contains the same node in added and removed sets")
	}
	if overlapFacts(addedFacts, removedFacts) {
		return Delta{}, fmt.Errorf("delta contains the same fact in added and removed sets")
	}
	return Delta{
		AddedNodes: addedNodes, RemovedNodes: removedNodes,
		AddedFacts: addedFacts, RemovedFacts: removedFacts,
	}, nil
}

// Normalize validates and canonicalizes the delta in place.
func (d *Delta) Normalize() error {
	if d == nil {
		return fmt.Errorf("cannot normalize a nil delta")
	}
	normalized, err := d.Normalized()
	if err != nil {
		return err
	}
	*d = normalized
	return nil
}

// IsEmpty reports whether the delta changes no semantic content.
func (d Delta) IsEmpty() bool {
	return len(d.AddedNodes) == 0 && len(d.RemovedNodes) == 0 && len(d.AddedFacts) == 0 && len(d.RemovedFacts) == 0
}

func normalizeNodes(nodes []Node, label string) ([]Node, error) {
	if len(nodes) == 0 {
		return nil, nil
	}
	result := make([]Node, len(nodes))
	seen := make(map[string]string, len(nodes))
	for i, node := range nodes {
		id, err := normalizeToken("node ID", node.ID)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		kind, err := normalizeToken("node kind", node.Kind)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		if previous, exists := seen[id]; exists && previous != kind {
			return nil, fmt.Errorf("%s: node %q has conflicting kinds %q and %q", label, id, previous, kind)
		}
		seen[id] = kind
		result[i] = Node{ID: id, Kind: kind}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].ID != result[j].ID {
			return result[i].ID < result[j].ID
		}
		return result[i].Kind < result[j].Kind
	})
	return uniqueNodes(result), nil
}

func normalizeFacts(facts []Fact, label string) ([]Fact, error) {
	if len(facts) == 0 {
		return nil, nil
	}
	result := make([]Fact, len(facts))
	for i, fact := range facts {
		subject, err := normalizeToken("fact subject", fact.Subject)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		predicate, err := normalizeToken("fact predicate", fact.Predicate)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		object, err := normalizeToken("fact object", fact.Object)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		result[i] = Fact{Subject: subject, Predicate: predicate, Object: object}
	}
	sort.Slice(result, func(i, j int) bool { return factLess(result[i], result[j]) })
	return uniqueFacts(result), nil
}

func normalizeValues(label string, values []string) ([]string, error) {
	if len(values) == 0 {
		return nil, nil
	}
	result := make([]string, len(values))
	for i, value := range values {
		normalized, err := normalizeToken(label, value)
		if err != nil {
			return nil, fmt.Errorf("%s[%d]: %w", label, i, err)
		}
		result[i] = normalized
	}
	sort.Strings(result)
	return uniqueStrings(result), nil
}

func normalizeToken(label, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%s is empty", label)
	}
	if strings.IndexFunc(value, func(r rune) bool { return r == '\n' || r == '\r' || r == '\t' || r == ' ' }) >= 0 {
		return "", fmt.Errorf("%s %q contains whitespace", label, value)
	}
	return value, nil
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func uniqueNodes(nodes []Node) []Node {
	result := make([]Node, 0, len(nodes))
	for _, node := range nodes {
		if len(result) == 0 || result[len(result)-1] != node {
			result = append(result, node)
		}
	}
	return result
}

func uniqueFacts(facts []Fact) []Fact {
	result := make([]Fact, 0, len(facts))
	for _, fact := range facts {
		if len(result) == 0 || result[len(result)-1] != fact {
			result = append(result, fact)
		}
	}
	return result
}

func nodeKey(node Node) string { return node.ID + "\x00" + node.Kind }

func factKey(fact Fact) string {
	return fact.Subject + "\x00" + fact.Predicate + "\x00" + fact.Object
}

func factLess(left, right Fact) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}

func overlapNodes(left, right []Node) bool {
	seen := make(map[string]struct{}, len(left))
	for _, node := range left {
		seen[nodeKey(node)] = struct{}{}
	}
	for _, node := range right {
		if _, exists := seen[nodeKey(node)]; exists {
			return true
		}
	}
	return false
}

func overlapFacts(left, right []Fact) bool {
	seen := make(map[string]struct{}, len(left))
	for _, fact := range left {
		seen[factKey(fact)] = struct{}{}
	}
	for _, fact := range right {
		if _, exists := seen[factKey(fact)]; exists {
			return true
		}
	}
	return false
}
