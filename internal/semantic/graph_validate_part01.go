package semantic

import (
	"fmt"
	"slices"
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
	slices.Sort(nodeIDs)
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
