package bodyexecution

import (
	"fmt"
	"sort"
)

func validateLocalObservations(traces []CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite, reads []CompositionReadSite, constructions []CompositionConstructionSite) (*CompositionLocalSummary, error) {
	if len(traces) < 1 || len(traces) > 128 || len(sites) > 4096 || len(uses) > 65536 {
		return nil, fmt.Errorf("local observation site or case bound differs")
	}
	if err := validateLocalSiteContract(sites, uses); err != nil {
		return nil, err
	}
	s := &CompositionLocalSummary{Schema: "gooo/native-local-summary/v1", Status: "COMPLETE",
		Scope: "authored scalar assignments and evaluated uses linked by lexical variable and function invocation; constructor element windows include evaluated helper uses; original use coordinates and complete causal dependencies are not inferred"}
	for _, t := range traces {
		if err := validateLocalTrace(t, sites, uses, reads, constructions); err != nil {
			return nil, fmt.Errorf("local case %d: %w", t.CaseIndex, err)
		}
		s.RecordedWrites += len(t.LocalWrites)
		s.RecordedUses += len(t.LocalUses)
		s.ElementWindows += len(t.LocalElements)
		for _, count := range []struct {
			target *int
			value  int
		}{{&s.Frames, t.LocalFrames}, {&s.FramesDropped, t.LocalFrameDropped}, {&s.WritesDropped, t.LocalWriteDropped}, {&s.UsesDropped, t.LocalUseDropped}, {&s.ReadsDropped, t.ReadsDropped}} {
			if err := addRecordedOriginDrops(count.target, count.value); err != nil {
				return nil, err
			}
		}
		for _, w := range t.LocalWrites {
			if !w.Completed {
				s.Interrupted++
			}
		}
		for _, u := range t.LocalUses {
			if u.Assignment < 0 {
				s.Unresolved++
			}
		}
	}
	if s.Frames > 65536 || s.RecordedWrites > 4096 || s.RecordedUses > 65536 || s.ElementWindows > 65536 ||
		s.FramesDropped > 0 && s.Frames != 65536 || s.WritesDropped > 0 && s.RecordedWrites != 4096 && s.FramesDropped == 0 ||
		s.UsesDropped > 0 && s.RecordedUses != 65536 && s.FramesDropped == 0 {
		return nil, fmt.Errorf("local observation population or dropped-event limit differs")
	}
	if s.FramesDropped > 0 || s.WritesDropped > 0 || s.UsesDropped > 0 || s.ReadsDropped > 0 || s.Unresolved > 0 || s.Interrupted > 0 {
		s.Status = "PARTIAL"
	}
	return s, nil
}

type localBindingKey struct{ frame, variable int }
type localFrameOwner struct{ root, activity string }

func validateLocalTrace(t CompositionTrace, sites []CompositionLocalSite, uses []CompositionLocalUseSite, reads []CompositionReadSite, constructions []CompositionConstructionSite) error {
	owners := map[int]localFrameOwner{}
	deliveries := map[string]CompositionDelivery{}
	for _, d := range t.Deliveries {
		deliveries[d.ActivityID] = d
	}
	bind := func(frame int, root, activity string) error {
		d, found := deliveries[root]
		if !found || len(d.BlockedBy) > 0 || frame < 0 || frame >= t.LocalFrames {
			return fmt.Errorf("local frame differs from active delivery")
		}
		owner := localFrameOwner{root, activity}
		if prior, found := owners[frame]; found && prior != owner {
			return fmt.Errorf("local frame crosses activity identity")
		}
		owners[frame] = owner
		return nil
	}
	completed := []int{}
	lastRead, lastUse := -1, -1
	for index, w := range t.LocalWrites {
		if w.SiteIndex < 0 || w.SiteIndex >= len(sites) || w.ReadStart < 0 || w.ReadEnd < w.ReadStart || w.ReadEnd > len(t.Reads) ||
			w.UseStart < 0 || w.UseEnd < w.UseStart || w.UseEnd > len(t.LocalUses) || w.ReadsDropped < 0 || w.ReadsDropped > t.ReadsDropped ||
			w.UsesDropped < 0 || w.UsesDropped > t.LocalUseDropped || w.ReadStart < lastRead || w.UseStart < lastUse {
			return fmt.Errorf("local assignment coordinates differ")
		}
		site := sites[w.SiteIndex]
		if err := bind(w.Frame, site.RootActivityID, site.ActivityID); err != nil {
			return err
		}
		if !w.Completed && deliveries[site.RootActivityID].Fault == nil {
			return fmt.Errorf("interrupted assignment has no activity fault")
		}
		if w.Completed {
			completed = append(completed, index)
		}
		lastRead, lastUse = w.ReadStart, w.UseStart
	}
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
	return validateLocalWindows(t, sites, uses, reads, constructions)
}
