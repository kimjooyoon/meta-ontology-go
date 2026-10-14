package bodyexecution

import "fmt"

func validateBridgeObservations(traces []CompositionTrace, r CompositionRuntime) (*CompositionBridgeSummary, error) {
	s := &CompositionBridgeSummary{Schema: "gooo/native-field-local-bridge-summary/v1", Status: "COMPLETE",
		MetadataSHA256: compositionDigest([]any{r.FieldProjectionSHA256, r.FieldBindings, r.FieldSites, r.FieldUseSites, r.LocalSites, r.LocalUseSites}),
		Scope:          "evaluated local/field read windows joined by actual case-relative assignments and shared completion clocks; no scalar values copied; control dependencies, precise helper-parameter selection and unobserved dependencies remain outside this journal profile"}
	if len(traces) < 1 || len(traces) > 128 || r.FieldObservation == nil {
		return nil, fmt.Errorf("bridge needs complete original cases and field observations")
	}
	for c, t := range traces {
		if t.CaseIndex != c {
			return nil, fmt.Errorf("bridge requires the complete original case population")
		}
		if err := validateBridgeTrace(t, r); err != nil {
			return nil, fmt.Errorf("bridge case%d: %w", c, err)
		}
		b := t.Bridge
		s.FieldWindows += len(b.Fields)
		s.LocalWindows += len(b.Locals)
		s.FieldUses += len(b.FieldClocks)
		s.LocalUses += len(b.LocalClocks)
		s.Completions += b.Completions
		for _, windows := range [][]CompositionBridgeWindow{b.Fields, b.Locals} {
			for _, w := range windows {
				s.CrossReads += w.OtherEnd - w.OtherStart
				if w.Completion < 0 {
					s.Interrupted++
				}
				if w.OtherDrops > 0 {
					s.Status = "PARTIAL"
				}
			}
		}
	}
	if s.LocalWindows > 4096 || s.FieldWindows > 4096 || s.LocalUses > 65536 || s.FieldUses > 65536 {
		return nil, fmt.Errorf("bridge retained populations exceed their original journal limits")
	}
	if s.Interrupted > 0 || r.FieldObservation.Status != "COMPLETE" || r.LocalObservation != nil && r.LocalObservation.Status != "COMPLETE" {
		s.Status = "PARTIAL"
	}
	return s, nil
}
