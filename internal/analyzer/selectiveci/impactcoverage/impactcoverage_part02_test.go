package impactcoverage

import (
	"reflect"
	"testing"
)

func literalCases(t *testing.T) []fixtureCase {
	t.Helper()
	a := boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a")
	b := boundSource("pkg/b.go", "b-1", "urn:gooo:entity:b")
	c := boundSource("pkg/c.go", "c-1", "urn:gooo:entity:c")
	return []fixtureCase{
		{"exact-replay", snap(t, "map", "reg", a), snap(t, "map", "reg", a), wantVector{
			DecisionExact, ReasonNoChange, false, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
		{"no-change-order-permutation", snap(t, "map", "reg", a, b), snap(t, "map", "reg", b, a), wantVector{
			DecisionExact, ReasonNoChange, false, 0, 0, 0, 0, 6, []string{}, []string{},
		}},
		{"modify-add-delete", snap(t, "map", "reg", a, b), snap(t, "map", "reg",
			boundSource("pkg/a.go", "a-2", "urn:gooo:entity:a"), c), wantVector{
			DecisionExact, ReasonComplete, false, 3, 3, 0, 3, 7,
			[]string{"urn:gooo:entity:a", "urn:gooo:entity:b", "urn:gooo:entity:c"}, []string{},
		}},
		{"relocation", snap(t, "map", "reg", a), snap(t, "map", "reg",
			boundSource("pkg/new.go", "a-1", "urn:gooo:entity:a")), wantVector{
			DecisionExact, ReasonComplete, false, 2, 2, 0, 1, 4, []string{"urn:gooo:entity:a"}, []string{},
		}},
		{"binding-set-only", snap(t, "map", "reg", a), snap(t, "map", "reg",
			boundSource("pkg/a.go", "a-1", "urn:gooo:entity:a", "urn:gooo:entity:b")), wantVector{
			DecisionExact, ReasonComplete, false, 1, 1, 0, 2, 4,
			[]string{"urn:gooo:entity:a", "urn:gooo:entity:b"}, []string{},
		}},
		{"unbound-changed-path", snap(t, "map", "reg", emptySource("pkg/u.go", "u-1")),
			snap(t, "map", "reg", emptySource("pkg/u.go", "u-2")), wantVector{
				DecisionUnknown, ReasonMissingBinding, true, 1, 0, 1, 0, 1, []string{}, []string{"pkg/u.go"},
			}},
		{"one-side-binding", snap(t, "map", "reg", a), snap(t, "map", "reg", emptySource("pkg/a.go", "a-2")), wantVector{
			DecisionExact, ReasonComplete, false, 1, 1, 0, 1, 2, []string{"urn:gooo:entity:a"}, []string{},
		}},
		{"registry-drift", snap(t, "map", "reg-a", a), snap(t, "map", "reg-b", a), wantVector{
			DecisionUnknown, ReasonAuthorityDrift, true, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
		{"source-map-drift", snap(t, "map-a", "reg", a), snap(t, "map-b", "reg", a), wantVector{
			DecisionUnknown, ReasonAuthorityDrift, true, 0, 0, 0, 0, 3, []string{}, []string{},
		}},
	}
}
func assertVector(t *testing.T, got Result, want wantVector) {
	t.Helper()
	if got.Decision != want.decision || got.Reason != want.reason || got.FullSuiteRequired != want.full {
		t.Fatalf("decision vector = %#v, want %s/%s/%t", got, want.decision, want.reason, want.full)
	}
	if got.ChangedBlobCount != want.changed || got.CoveredChangedBlobCount != want.covered ||
		got.UncoveredChangedBlobCount != want.open || got.ChangedBindingCount != want.bindings ||
		got.DeterministicWorkUnits != want.work {
		t.Fatalf("numeric vector = %#v, want changed=%d covered=%d open=%d bindings=%d work=%d",
			got, want.changed, want.covered, want.open, want.bindings, want.work)
	}
	if !reflect.DeepEqual(got.ChangedStableIDs, want.ids) || !reflect.DeepEqual(got.UncoveredPaths, want.paths) {
		t.Fatalf("set vector = %#v/%#v, want %#v/%#v", got.ChangedStableIDs, got.UncoveredPaths, want.ids, want.paths)
	}
}
