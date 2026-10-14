package bodyexecution

import (
	"fmt"
	"sort"
)

func validateBridgeTrace(t CompositionTrace, r CompositionRuntime) error {
	b := t.Bridge
	if b == nil || len(b.Locals) != len(t.LocalWrites) || len(b.Fields) != len(t.FieldWrites) || len(b.LocalClocks) != len(t.LocalUses) || len(b.FieldClocks) != len(t.FieldUses) || b.Completions < 0 || b.Completions > len(b.Locals)+len(b.Fields) {
		return fmt.Errorf("bridge window, read-clock or completion population differs")
	}
	seen := make([]bool, b.Completions)
	fieldCount := 0
	for i, w := range t.FieldWrites {
		if err := validateBridgeWindow(b.Fields[i], w.Completed, len(t.LocalUses), t.LocalUseDropped, b.Completions, seen); err != nil {
			return err
		}
		if w.Completed {
			fieldCount++
		}
	}
	for i, w := range t.LocalWrites {
		if err := validateBridgeWindow(b.Locals[i], w.Completed, len(t.FieldUses), t.FieldUseDropped, b.Completions, seen); err != nil {
			return err
		}
	}
	for _, found := range seen {
		if !found {
			return fmt.Errorf("shared bridge completion sequence has a gap")
		}
	}
	fieldOrder, err := bridgeFieldCompletionOrder(t, fieldCount)
	if err != nil {
		return err
	}
	if err := validateBridgeFieldUses(t, fieldOrder); err != nil {
		return err
	}
	if err := validateBridgeLocalUses(t, r.LocalSites, r.LocalUseSites); err != nil {
		return err
	}
	return validateBridgeEvaluationWindows(t, r)
}

func validateBridgeWindow(w CompositionBridgeWindow, completed bool, otherUses, otherDrops, completions int, seen []bool) error {
	if w.OtherStart < 0 || w.OtherEnd < w.OtherStart || w.OtherEnd > otherUses || w.OtherDrops < 0 || w.OtherDrops > otherDrops || w.StartClock < 0 || w.StartClock > completions {
		return fmt.Errorf("bridge window bounds, losses or start clock differ")
	}
	if !completed {
		if w.Completion != -1 {
			return fmt.Errorf("unfinished write claims a shared completion")
		}
		return nil
	}
	if w.Completion < w.StartClock || w.Completion < 0 || w.Completion >= completions || seen[w.Completion] {
		return fmt.Errorf("shared bridge completion is duplicated, outside bounds or before evaluation")
	}
	seen[w.Completion] = true
	return nil
}

func bridgeFieldCompletionOrder(t CompositionTrace, count int) ([]int, error) {
	order := make([]int, count)
	for i, w := range t.FieldWrites {
		if !w.Completed {
			continue
		}
		if w.Completion < 0 || w.Completion >= count {
			return nil, fmt.Errorf("field completion differs from bridge population")
		}
		order[w.Completion] = t.Bridge.Fields[i].Completion
		if w.ParentAssignment >= 0 && (w.ParentAssignment >= len(t.FieldWrites) || t.Bridge.Fields[w.ParentAssignment].Completion < 0 || t.Bridge.Fields[w.ParentAssignment].Completion >= order[w.Completion]) {
			return nil, fmt.Errorf("field parent follows its child in the shared clock")
		}
	}
	for i := 1; i < len(order); i++ {
		if order[i] <= order[i-1] {
			return nil, fmt.Errorf("shared completion order contradicts the original field clock")
		}
	}
	return order, nil
}

func validateBridgeFieldUses(t CompositionTrace, order []int) error {
	previous := 0
	for i, u := range t.FieldUses {
		clock := t.Bridge.FieldClocks[i]
		if clock < previous || clock > t.Bridge.Completions || sort.SearchInts(order, clock) != u.CompletionClock {
			return fmt.Errorf("bridge field read clock contradicts original field completion order")
		}
		if u.Assignment >= 0 && (u.Assignment >= len(t.Bridge.Fields) || t.Bridge.Fields[u.Assignment].Completion < 0 || t.Bridge.Fields[u.Assignment].Completion >= clock) {
			return fmt.Errorf("bridge field read precedes its assigned completion")
		}
		previous = clock
	}
	return nil
}
