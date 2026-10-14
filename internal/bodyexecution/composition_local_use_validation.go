package bodyexecution

import (
	"fmt"
	"sort"
)

func validateLocalUses(t CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite, completed []int, bind func(int, string, string) error) error {
	// A write is reserved before its RHS. Its completion boundary identifies the
	// latest finished assignment at each evaluated use, including reassignment.
	sort.SliceStable(completed, func(i, j int) bool { return t.LocalWrites[completed[i]].UseEnd < t.LocalWrites[completed[j]].UseEnd })
	latest := map[localBindingKey]int{}
	cursor := 0
	for index, u := range t.LocalUses {
		if u.SiteIndex < 0 || u.SiteIndex >= len(uses) {
			return fmt.Errorf("local use site differs")
		}
		site := uses[u.SiteIndex]
		if err := bind(u.Frame, site.RootActivityID, site.ActivityID); err != nil {
			return err
		}
		for cursor < len(completed) && t.LocalWrites[completed[cursor]].UseEnd <= index {
			id := completed[cursor]
			w := t.LocalWrites[id]
			latest[localBindingKey{w.Frame, sites[w.SiteIndex].VariableOrdinal}] = id
			cursor++
		}
		if u.Assignment < 0 {
			if u.Assignment == -2 && u.Reason == "WRITE_LIMIT" && t.LocalWriteDropped > 0 {
				continue
			}
			if u.Assignment == -1 && u.Reason == "NO_RECORDED_ASSIGNMENT" {
				if _, found := latest[localBindingKey{u.Frame, site.VariableOrdinal}]; !found {
					continue
				}
			}
			return fmt.Errorf("unresolved local use has no matching reason")
		}
		if u.Assignment >= len(t.LocalWrites) || u.Reason != "" {
			return fmt.Errorf("local use assignment reference differs")
		}
		w := t.LocalWrites[u.Assignment]
		origin := sites[w.SiteIndex]
		last, found := latest[localBindingKey{u.Frame, site.VariableOrdinal}]
		if !found || last != u.Assignment || !w.Completed || w.UseEnd > index || w.Frame != u.Frame || origin.VariableOrdinal != site.VariableOrdinal || origin.ScalarType != site.ScalarType {
			return fmt.Errorf("local use does not refer to the latest completed lexical assignment")
		}
	}
	return nil
}
