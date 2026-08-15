package semantic

import (
	"fmt"
	"sort"
)

func (g Graph) Validate() error {
	issues := &ValidationErrors{}
	validateNodes(g, issues)
	validateFactMap(g, g.facts, FactDeterministic, issues)
	validateFactMap(g, g.candidates, FactCandidate, issues)
	for key := range g.facts {
		if _, exists := g.candidates[key]; exists {
			issues.add("fact-overlap", "a deterministic fact cannot coexist with its candidate", key.Subject, key.Object)
		}
	}
	if len(issues.Issues) == 0 {
		return nil
	}
	return issues
}

func validateNodes(g Graph, issues *ValidationErrors) {
	nameOwners := make(map[NameRef]ID, len(g.names))
	identityOwners := make(map[ID]graphIdentityOwner)
	nodeIDs := make([]ID, 0, len(g.nodes))
	for id := range g.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool { return nodeIDs[i] < nodeIDs[j] })
	for _, id := range nodeIDs {
		node := g.nodes[id]
		normalized, err := node.Normalized()
		if err != nil {
			issues.add("node", err.Error(), node.ID, "")
			continue
		}
		if normalized.ID != id {
			issues.add("node-key", "node map key does not match normalized node ID", id, normalized.ID)
		}
		registerGraphIdentity(identityOwners, normalized.ID, graphIdentityOwner{kind: "node", parent: normalized.ID}, issues)
		for _, ref := range nodeNameRefs(normalized) {
			if owner, exists := nameOwners[ref]; exists && owner != normalized.ID {
				issues.add("name-collision", fmt.Sprintf("%s/%s belongs to %s and %s", ref.Namespace, ref.Name, owner, normalized.ID), owner, normalized.ID)
			}
			nameOwners[ref] = normalized.ID
		}
		for _, field := range normalized.Fields {
			registerGraphIdentity(identityOwners, field.ID, graphIdentityOwner{kind: "field", parent: normalized.ID}, issues)
		}
	}
	validateNameIndex(g, nameOwners, issues)
}

type graphIdentityOwner struct {
	kind   string
	parent ID
}

func registerGraphIdentity(owners map[ID]graphIdentityOwner, id ID, incoming graphIdentityOwner, issues *ValidationErrors) {
	if owner, exists := owners[id]; exists {
		code := "id-collision"
		if owner.kind == "field" && incoming.kind == "field" {
			code = "field-id-collision"
		}
		issues.add(code, fmt.Sprintf("%s %s on %s collides with %s on %s", incoming.kind, id, incoming.parent, owner.kind, owner.parent), incoming.parent, owner.parent)
	}
	owners[id] = incoming
}

func validateNameIndex(g Graph, expected map[NameRef]ID, issues *ValidationErrors) {
	actualRefs := make([]NameRef, 0, len(g.names))
	for ref := range g.names {
		actualRefs = append(actualRefs, ref)
	}
	sort.Slice(actualRefs, func(i, j int) bool { return nameRefLess(actualRefs[i], actualRefs[j]) })
	for _, ref := range actualRefs {
		actual := g.names[ref]
		want, exists := expected[ref]
		if !exists {
			issues.add("name-index-stale", fmt.Sprintf("%s/%s has no declared node", ref.Namespace, ref.Name), actual, "")
			continue
		}
		if actual != want {
			issues.add("name-index-owner", fmt.Sprintf("%s/%s points to %s, want %s", ref.Namespace, ref.Name, actual, want), actual, want)
		}
	}

	missingRefs := make([]NameRef, 0, len(expected))
	for ref := range expected {
		if _, exists := g.names[ref]; !exists {
			missingRefs = append(missingRefs, ref)
		}
	}
	sort.Slice(missingRefs, func(i, j int) bool { return nameRefLess(missingRefs[i], missingRefs[j]) })
	for _, ref := range missingRefs {
		issues.add("name-index-missing", fmt.Sprintf("%s/%s is not indexed", ref.Namespace, ref.Name), expected[ref], "")
	}
}

func nameRefLess(left, right NameRef) bool {
	if left.Namespace != right.Namespace {
		return left.Namespace < right.Namespace
	}
	return left.Name < right.Name
}

func validateFactMap(g Graph, facts map[FactKey]Fact, expected FactStatus, issues *ValidationErrors) {
	keys := make([]FactKey, 0, len(facts))
	for key := range facts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return factKeyLess(keys[i], keys[j]) })
	for _, key := range keys {
		validateStoredFact(g, key, facts[key], expected, issues)
	}
}

func validateStoredFact(g Graph, key FactKey, fact Fact, expected FactStatus, issues *ValidationErrors) {
	normalized, err := fact.Normalized()
	if err != nil {
		issues.add("fact", err.Error(), fact.Subject, fact.Object)
		return
	}
	if normalized.Status != expected {
		issues.add("fact-status", fmt.Sprintf("fact is stored as %s but marked %s", expected, normalized.Status), fact.Subject, fact.Object)
	}
	if normalized.Key() != key {
		issues.add("fact-key", "fact key is not normalized", fact.Subject, fact.Object)
	}
	subject, subjectOK := g.nodes[normalized.Subject]
	object, objectOK := g.nodes[normalized.Object]
	if !subjectOK {
		issues.add("missing-subject", "fact subject is not declared", normalized.Subject, normalized.Object)
	}
	if !objectOK {
		issues.add("missing-object", "fact object is not declared", normalized.Subject, normalized.Object)
	}
	if subjectOK && objectOK {
		if err := normalized.Predicate.ValidateKinds(subject.Kind, object.Kind); err != nil {
			issues.add("relation-kind", err.Error(), normalized.Subject, normalized.Object)
		}
	}
}

func factKeyLess(left, right FactKey) bool {
	if left.Subject != right.Subject {
		return left.Subject < right.Subject
	}
	if left.Predicate != right.Predicate {
		return left.Predicate < right.Predicate
	}
	return left.Object < right.Object
}

func (g Graph) Normalized() (Graph, error) {
	out := NewGraph()
	for _, node := range g.Nodes() {
		if err := out.AddNode(node); err != nil {
			return Graph{}, err
		}
	}
	for _, fact := range g.AllFacts() {
		if err := out.AddFact(fact); err != nil {
			return Graph{}, err
		}
	}
	if err := out.Validate(); err != nil {
		return Graph{}, err
	}
	return out, nil
}

func (g *Graph) Normalize() error {
	normalized, err := g.Normalized()
	if err != nil {
		return err
	}
	*g = normalized
	return nil
}

func (g Graph) Clone() Graph {
	clone := NewGraph()
	for _, node := range g.Nodes() {
		if err := clone.AddNode(node); err != nil {
			return NewGraph()
		}
	}
	for _, fact := range g.AllFacts() {
		if err := clone.AddFact(fact); err != nil {
			return NewGraph()
		}
	}
	return clone
}
