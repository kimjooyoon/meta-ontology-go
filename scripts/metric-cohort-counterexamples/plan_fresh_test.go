package main

import "testing"

func TestCountercaseDeclarationFresh1681(t *testing.T) {
	planned := countercases()
	if len(planned) != 16 {
		t.Fatal("the declared native case denominator changed")
	}
	seen := map[string]bool{}
	for _, row := range planned {
		if row.ID == "" || row.Kind == "" || row.Reason == "" || seen[row.ID] {
			t.Fatal("countercase identity or expected refusal is missing or duplicated")
		}
		seen[row.ID] = true
		if row.Kind != "PRODUCER_INPUT_MUTATION" && row.Kind != "RESEALED_RECEIPT_MUTATION" &&
			row.Kind != "SOURCE_INPUT_MUTATION" && row.Kind != "MISSING_INPUT" {
			t.Fatal("unimplemented native case kind")
		}
	}
}
