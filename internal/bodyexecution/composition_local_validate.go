package bodyexecution

import (
	"fmt"
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
	completed, err := validateLocalWrites(t, sites, deliveries, bind)
	if err != nil {
		return err
	}
	if err := validateLocalUses(t, sites, uses, completed, bind); err != nil {
		return err
	}
	return validateLocalWindows(t, sites, uses, reads, constructions)
}
