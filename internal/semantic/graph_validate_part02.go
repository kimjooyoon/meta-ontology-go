package semantic

import (
	"fmt"
	"slices"
	"sort"
)

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

// validateFieldOwnerIndex derives its expected owners from stored nodes rather
// than trusting the lookup used during insertion.
func validateFieldOwnerIndex(g Graph, expected map[ID]graphIdentityOwner, issues *ValidationErrors) {
	actualIDs := make([]ID, 0, len(g.fieldOwners))
	for id := range g.fieldOwners {
		actualIDs = append(actualIDs, id)
	}
	slices.Sort(actualIDs)
	for _, id := range actualIDs {
		actual := g.fieldOwners[id]
		want, exists := expected[id]
		if !exists || want.kind != "field" {
			issues.add("field-index-stale", fmt.Sprintf("%s has no declared field", id), actual, "")
			continue
		}
		if actual != want.parent {
			message := fmt.Sprintf("%s points to %s, want %s", id, actual, want.parent)
			issues.add("field-index-owner", message, actual, want.parent)
		}
	}
	missingIDs := make([]ID, 0)
	for id, owner := range expected {
		if owner.kind == "field" {
			if _, exists := g.fieldOwners[id]; !exists {
				missingIDs = append(missingIDs, id)
			}
		}
	}
	slices.Sort(missingIDs)
	for _, id := range missingIDs {
		issues.add("field-index-missing", fmt.Sprintf("%s is not indexed", id), expected[id].parent, "")
	}
}
