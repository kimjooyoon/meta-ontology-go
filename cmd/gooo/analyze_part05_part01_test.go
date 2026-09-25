package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCanonicalFixPlanDiagnosticsOrdersObservableFields(t *testing.T) {
	span := fixPlanSpan{
		File:  "fixture.gooo",
		Start: fixPlanPosition{Offset: 4, Line: 2, Column: 5},
		End:   fixPlanPosition{Offset: 8, Line: 2, Column: 9},
	}
	input := []fixPlanDiagnostic{
		{
			RepairID: "repair-warning", Phase: "semantic", Severity: "warning",
			Code: "semantic.same", Message: "same", Span: span,
			Applicability: "potential", Status: "deferred",
		},
		{
			RepairID: "repair-error", Phase: "semantic", Severity: "error",
			Code: "semantic.same", Message: "same", Span: span,
			Applicability: "not-evaluated", Status: "blocked",
		},
	}
	reversed := append([]fixPlanDiagnostic(nil), input...)
	reversed[0], reversed[1] = reversed[1], reversed[0]

	first, err := json.Marshal(canonicalFixPlanDiagnostics(input))
	if err != nil {
		t.Fatal(err)
	}
	second, err := json.Marshal(canonicalFixPlanDiagnostics(reversed))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("fix-plan diagnostics changed with insertion order:\nfirst=%s\nsecond=%s", first, second)
	}
	ordered := canonicalFixPlanDiagnostics(input)
	if ordered[0].Severity != "error" || ordered[1].Severity != "warning" {
		t.Fatalf("fix-plan diagnostic severity order = %#v", ordered)
	}
}
