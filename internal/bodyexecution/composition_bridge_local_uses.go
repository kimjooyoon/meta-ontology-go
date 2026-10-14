package bodyexecution

import (
	"fmt"
	"slices"
	"sort"
)

func validateBridgeLocalUses(t CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite) error {
	groups := map[localBindingKey][]int{}
	for i, w := range t.LocalWrites {
		if w.SiteIndex < 0 || w.SiteIndex >= len(sites) {
			return fmt.Errorf("bridge local write has no source site")
		}
		if w.Completed {
			key := localBindingKey{w.Frame, sites[w.SiteIndex].VariableOrdinal}
			groups[key] = append(groups[key], i)
		}
	}
	for _, group := range groups {
		slices.SortFunc(group, func(a, b int) int { return t.Bridge.Locals[a].Completion - t.Bridge.Locals[b].Completion })
	}
	previous := 0
	for i, u := range t.LocalUses {
		if u.SiteIndex < 0 || u.SiteIndex >= len(uses) {
			return fmt.Errorf("bridge local read has no source site")
		}
		clock := t.Bridge.LocalClocks[i]
		if clock < previous || clock > t.Bridge.Completions {
			return fmt.Errorf("bridge local read clock is not increasing or is outside the completion population")
		}
		previous = clock
		group := groups[localBindingKey{u.Frame, uses[u.SiteIndex].VariableOrdinal}]
		position := sort.Search(len(group), func(j int) bool { return t.Bridge.Locals[group[j]].Completion >= clock })
		expected := -1
		if position > 0 {
			expected = group[position-1]
		}
		if u.Assignment >= 0 && expected != u.Assignment || u.Assignment == -1 && expected >= 0 {
			return fmt.Errorf("bridge local read does not reference its latest completed lexical assignment")
		}
	}
	return nil
}
