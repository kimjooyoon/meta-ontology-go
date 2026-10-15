package bodyexecution

import "fmt"

func validateBridgeEvaluationWindows(t CompositionTrace, r CompositionRuntime) error {
	localOwners := make([]string, len(t.LocalUses))
	for i, u := range t.LocalUses {
		if u.SiteIndex < 0 || u.SiteIndex >= len(r.LocalUseSites) {
			return fmt.Errorf("bridge local window read has no source site")
		}
		localOwners[i] = r.LocalUseSites[u.SiteIndex].RootActivityID
	}
	fieldOwners := make([]string, len(t.FieldUses))
	for i, u := range t.FieldUses {
		if u.SiteIndex < 0 || u.SiteIndex >= len(r.FieldUseSites) {
			return fmt.Errorf("bridge field window read has no source site")
		}
		site := r.FieldUseSites[u.SiteIndex]
		if site.Binding < 0 || site.Binding >= len(r.FieldBindings) {
			return fmt.Errorf("bridge field window read has no binding")
		}
		fieldOwners[i] = r.FieldBindings[site.Binding].RootActivityID
	}
	if err := validateBridgeFieldWindows(t, r, fieldOwners, localOwners); err != nil {
		return err
	}
	return validateBridgeLocalWindows(t, r, fieldOwners, localOwners)
}

func validateBridgeFieldWindows(t CompositionTrace, r CompositionRuntime, fieldOwners, localOwners []string) error {
	for i, w := range t.FieldWrites {
		if w.SiteIndex < 0 || w.SiteIndex >= len(r.FieldSites) {
			return fmt.Errorf("bridge field write has no source site")
		}
		site := r.FieldSites[w.SiteIndex]
		if site.Binding < 0 || site.Binding >= len(r.FieldBindings) {
			return fmt.Errorf("bridge field write has no binding")
		}
		root, window := r.FieldBindings[site.Binding].RootActivityID, t.Bridge.Fields[i]
		if err := validateBridgeReadWindow(window, w.UseStart, w.UseEnd, t.Bridge.FieldClocks, fieldOwners, root, t.Bridge.Completions); err != nil {
			return err
		}
		if err := validateBridgeReadWindow(window, window.OtherStart, window.OtherEnd, t.Bridge.LocalClocks, localOwners, root, t.Bridge.Completions); err != nil {
			return err
		}
	}
	return nil
}

func validateBridgeLocalWindows(t CompositionTrace, r CompositionRuntime, fieldOwners, localOwners []string) error {
	for i, w := range t.LocalWrites {
		if w.SiteIndex < 0 || w.SiteIndex >= len(r.LocalSites) {
			return fmt.Errorf("bridge local write has no source site")
		}
		root, window := r.LocalSites[w.SiteIndex].RootActivityID, t.Bridge.Locals[i]
		if err := validateBridgeReadWindow(window, w.UseStart, w.UseEnd, t.Bridge.LocalClocks, localOwners, root, t.Bridge.Completions); err != nil {
			return err
		}
		if err := validateBridgeReadWindow(window, window.OtherStart, window.OtherEnd, t.Bridge.FieldClocks, fieldOwners, root, t.Bridge.Completions); err != nil {
			return err
		}
	}
	return nil
}

func validateBridgeReadWindow(window CompositionBridgeWindow, start, end int, clocks []int, owners []string, root string, completions int) error {
	if start < 0 || end < start || end > len(clocks) || len(owners) != len(clocks) {
		return fmt.Errorf("bridge evaluation window crosses retained read bounds")
	}
	limit := window.Completion
	if limit < 0 {
		limit = completions
	}
	for j := start; j < end; j++ {
		if clocks[j] < window.StartClock || clocks[j] > limit || owners[j] != root {
			return fmt.Errorf("bridge window crosses evaluation clocks or source root identity")
		}
	}
	return nil
}
