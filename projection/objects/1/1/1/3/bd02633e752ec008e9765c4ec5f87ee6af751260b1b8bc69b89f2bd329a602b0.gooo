package impactgraph_test

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
	"reflect"
	"sort"
	"testing"
)

func assertResult(t *testing.T, got impactgraph.Result, want expectedResult) {
	t.Helper()
	if string(got.Decision) != want.decision {
		t.Errorf("Decision = %q, want %q", got.Decision, want.decision)
	}
	if got.FullSuiteRequired != want.fullSuiteRequired {
		t.Errorf("FullSuiteRequired = %v, want %v", got.FullSuiteRequired, want.fullSuiteRequired)
	}
	assertStringSet(t, "Required", got.Required, want.required)
	assertStringSet(t, "ExecutedRequired", got.ExecutedRequired, want.executedRequired)
	assertStringSet(t, "Missed", got.Missed, want.missed)
	assertStringSet(t, "Extra", got.Extra, want.extra)
	if got.Numerator != want.coverageNumerator {
		t.Errorf("Numerator = %d, want %d", got.Numerator, want.coverageNumerator)
	}
	if got.Denominator != want.coverageDenominator {
		t.Errorf("Denominator = %d, want %d", got.Denominator, want.coverageDenominator)
	}
}
func assertStringSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	actual := append([]string(nil), got...)
	expected := append([]string(nil), want...)
	sort.Strings(actual)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s = %#v, want %#v", label, actual, expected)
	}
}
