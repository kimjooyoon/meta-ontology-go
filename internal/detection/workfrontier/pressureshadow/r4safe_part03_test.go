package pressureshadow

import (
	"strings"
	"testing"
)

func TestR4SafeStrictPermutationK21AndOpaqueMutation(t *testing.T) {
	base := r4SafeInput(t, 10)
	raw, _ := CanonicalR4SafeInputBytes(base)
	for _, data := range [][]byte{
		[]byte(strings.Replace(string(raw), `"schema":`, `"expected_label":"PASS","schema":`, 1)),
		[]byte(strings.Replace(string(raw), `"schema":`, `"schema":"duplicate","schema":`, 1)),
		append(raw, []byte(`{}`)...),
		[]byte(strings.Replace(string(raw), `"path/root"`, `"path root"`, 1)),
	} {
		got := ValidateR4SafeBytes(data)
		if got.Decision != DecisionFailClosed || got.Reason != ReasonInvalidInput || !got.FullSuiteRequired {
			t.Fatalf("strict result = %#v", got)
		}
	}
	duplicateRow := base
	duplicateRow.PathCoverage = append(duplicateRow.PathCoverage, duplicateRow.PathCoverage[0])
	duplicatePressure := r4SafeInput(t, 10)
	conflict := duplicatePressure.PathCoverage[0].Coverage.PressureRecords[0]
	conflict.CategoryID = "conflicting-category"
	records := duplicatePressure.PathCoverage[0].Coverage.PressureRecords
	duplicatePressure.PathCoverage[0].Coverage.PressureRecords = append(records, conflict)
	for _, input := range []R4SafeInput{duplicateRow, duplicatePressure} {
		if got := ValidateR4Safe(input); got.Decision != DecisionFailClosed || got.Reason != ReasonInvalidInput {
			t.Fatalf("duplicate identity result = %#v", got)
		}
	}
}
