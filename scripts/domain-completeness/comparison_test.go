package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func TestComparisonKeepsUnknownDeltasNull(t *testing.T) {
	for _, state := range []string{"UNKNOWN", "FAIL_CLOSED", "", "unsupported"} {
		for _, side := range []string{"prior", "current"} {
			t.Run(side+"-"+state, func(t *testing.T) {
				prior := Dimension{Status: "PROGRESS", Numerator: 1, Denominator: 4}
				current := Dimension{Status: "PROGRESS", Numerator: 3, Denominator: 4}
				if side == "prior" {
					prior.Status = state
				} else {
					current.Status = state
				}
				assertNullComparison(t, compareDimension(current, prior))
			})
		}
	}
}

func TestComparisonRequiresResolvedUnitsOnBothSides(t *testing.T) {
	for _, side := range []string{"prior", "current"} {
		for _, evidence := range []string{"unknown", "refuted"} {
			t.Run(side+"-"+evidence, func(t *testing.T) {
				prior := Dimension{Status: "PROGRESS", Numerator: 1, Denominator: 4}
				current := Dimension{Status: "PASS", Numerator: 4, Denominator: 4}
				changed := &prior
				if side == "current" {
					changed = &current
				}
				if evidence == "unknown" {
					changed.UnknownUnits = 1
				} else {
					changed.RefutedUnits = 1
				}
				assertNullComparison(t, compareDimension(current, prior))
			})
		}
	}
}

func assertNullComparison(t *testing.T, delta DimensionDelta) {
	t.Helper()
	if delta.Status != "UNKNOWN_UNRESOLVED_EVIDENCE" || delta.NumeratorDelta != nil {
		t.Fatalf("unresolved evidence gained a numeric delta: %#v", delta)
	}
	raw, err := json.Marshal(delta)
	if err != nil || !bytes.Contains(raw, []byte(`"numerator_delta":null`)) {
		t.Fatalf("unknown delta JSON = %s, %v", raw, err)
	}
}

func TestComparisonRetainsMeasuredImprovementStasisAndRegression(t *testing.T) {
	for _, counts := range [][2]int{{0, 2}, {0, 0}, {4, 1}} {
		prior := Dimension{Numerator: counts[0], Denominator: 4, Status: classify(counts[0], 4, 0, false)}
		current := Dimension{Numerator: counts[1], Denominator: 4, Status: classify(counts[1], 4, 0, false)}
		delta := compareDimension(current, prior)
		if delta.Status != "COMPARABLE" || delta.NumeratorDelta == nil || *delta.NumeratorDelta != counts[1]-counts[0] {
			t.Fatalf("measured %d -> %d comparison = %#v", counts[0], counts[1], delta)
		}
		raw, err := json.Marshal(delta)
		if err != nil || bytes.Contains(raw, []byte(`"numerator_delta":null`)) {
			t.Fatalf("measured delta lost its numeric representation: %s, %v", raw, err)
		}
	}
}

func TestComparisonRejectsInvalidCurrentCounts(t *testing.T) {
	for _, kind := range []string{"unknown", "refuted"} {
		baseline := observedZeroBaseline(t)
		current := observedZeroBaseline(t)
		current.SubjectSHA = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		current.Snapshot.SubjectSHA = current.SubjectSHA
		if kind == "unknown" {
			current.Dimensions[0].UnknownUnits = -1
		} else {
			current.Dimensions[0].RefutedUnits = -1
		}
		if got := compareReports(current, baseline, true); got.Status != "UNKNOWN_INCOMPATIBLE_BASELINE" || len(got.Dimensions) != 0 {
			t.Fatalf("negative %s count gained a comparison: %#v", kind, got)
		}
	}
}

func TestComparisonRuleChangesProfileSemanticIdentity(t *testing.T) {
	source, err := os.ReadFile("profile.gooo")
	if err != nil {
		t.Fatal(err)
	}
	changed := bytes.Replace(source, []byte("input0 == 0 &&"), []byte("input0 <= 0 &&"), 1)
	if bytes.Equal(source, changed) {
		t.Fatal("comparison rule fixture did not change")
	}
	before, err := compileProfile("profile.gooo", source)
	if err != nil {
		t.Fatal(err)
	}
	after, err := compileProfile("comparison-changed.gooo", changed)
	if err != nil {
		t.Fatal(err)
	}
	if before.SemanticHash == after.SemanticHash {
		t.Fatal("changed Gooo comparison rule retained the previous semantic identity")
	}
}
