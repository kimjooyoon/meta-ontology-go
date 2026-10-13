package bodyexecution

import "testing"

func localContractFixture10589() (CompositionRuntime, []CompositionTrace) {
	root := "urn:local10589:root"
	r := CompositionRuntime{
		LocalSites:        []CompositionLocalSite{{RootActivityID: root, ActivityID: root, Name: "value", ScalarType: "int64", Start: 0, End: 1}},
		LocalUseSites:     []CompositionLocalUseSite{{RootActivityID: root, ActivityID: root, Name: "value", ScalarType: "int64", Start: 2, End: 3}},
		ConstructionSites: []CompositionConstructionSite{{RootActivityID: root}},
	}
	traces := []CompositionTrace{{CaseIndex: 0, LocalFrames: 1, Deliveries: []CompositionDelivery{{ActivityID: root}},
		LocalWrites:   []CompositionLocalWrite{{SiteIndex: 0, Frame: 0, Completed: true}, {SiteIndex: 0, Frame: 0, Completed: true}},
		LocalUses:     []CompositionLocalUse{{SiteIndex: 0, Frame: 0, Assignment: 1}, {SiteIndex: 0, Frame: 0, Assignment: 1}},
		Constructions: []CompositionConstruction{{SiteIndex: 0, Completed: true, Elements: []CompositionConstructionElement{{Completed: true}, {Completed: true}}}},
		LocalElements: []CompositionLocalElementWindow{{Construction: 0, Element: 0, UseStart: 0, UseEnd: 1}, {Construction: 0, Element: 1, UseStart: 1, UseEnd: 2}},
	}}
	return r, traces
}

func TestLocalRecordedContract10589(t *testing.T) {
	check := func(r CompositionRuntime, traces []CompositionTrace) error {
		_, err := validateLocalObservations(traces, r.LocalSites, r.LocalUseSites, r.ReadSites, r.ConstructionSites)
		return err
	}
	r, traces := localContractFixture10589()
	if err := check(r, traces); err != nil {
		t.Fatal("fixture", err)
	}
	mutations := map[string]func(*CompositionRuntime, []CompositionTrace){
		"stale_assignment":           func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalUses[0].Assignment = 0 },
		"future_assignment":          func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalWrites[1].UseEnd = 1 },
		"different_invocation":       func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalUses[0].Frame = 1 },
		"unfinished_without_fault":   func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalWrites[1].Completed = false },
		"read_outside_window":        func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalWrites[0].ReadEnd = 1 },
		"duplicate_element":          func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalElements[1].Element = 0 },
		"missing_element":            func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalElements = t[0].LocalElements[:1] },
		"noncontiguous_element":      func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalElements[1].UseStart = 0 },
		"unknown_constructor":        func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalElements[0].Construction = 1 },
		"different_constructor_root": func(r *CompositionRuntime, _ []CompositionTrace) { r.ConstructionSites[0].RootActivityID = "urn:other" },
		"negative_drop":              func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalWriteDropped = -1 },
		"impossible_use_drop":        func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalUseDropped = 1 },
		"invalid_scalar_type":        func(r *CompositionRuntime, _ []CompositionTrace) { r.LocalUseSites[0].ScalarType = "float64" },
		"negative_lexical_identity": func(r *CompositionRuntime, _ []CompositionTrace) {
			r.LocalSites[0].VariableOrdinal = -1
			r.LocalUseSites[0].VariableOrdinal = -1
		},
		"different_declared_name": func(r *CompositionRuntime, _ []CompositionTrace) { r.LocalUseSites[0].Name = "other" },
		"empty_projection_range":  func(r *CompositionRuntime, _ []CompositionTrace) { r.LocalUseSites[0].End = 2 },
		"duplicate_static_write": func(r *CompositionRuntime, _ []CompositionTrace) {
			r.LocalSites = append(r.LocalSites, r.LocalSites[0])
		},
		"unresolved_without_reason": func(_ *CompositionRuntime, t []CompositionTrace) { t[0].LocalUses[0].Assignment = -1 },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			r, traces := localContractFixture10589()
			change(&r, traces)
			if err := check(r, traces); err == nil {
				t.Fatal("contradictory local origin accepted")
			}
		})
	}
	t.Log("one coherent local-origin fixture;18 contradictory binding, temporal, window and population variants rejected;no native execution")
}
