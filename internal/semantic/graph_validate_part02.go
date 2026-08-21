package semantic

import (
	"fmt"
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
