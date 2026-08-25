package main

import (
	"encoding/json"
	"testing"
)

func TestReduceLowersUnknownSourceReceipt(t *testing.T) {
	density, extraction := exactReports()
	density.Schema = "future"
	report := reduce("abc", density, extraction, nil, nil)
	if report.Decision != "FAIL_CLOSED" || report.Resolution != "LOWER_RESOLUTION" ||
		report.Reason != "DENSITY_RECEIPT_UNKNOWN" || report.Coordinates.SourceReceipts != 1 ||
		report.Coordinates.Unknowns != 1 {
		t.Fatalf("report=%#v", report)
	}
}

func TestReduceLowersUnknownDensityStatus(t *testing.T) {
	density, extraction := exactReports()
	density.Subjects[0].Status = "future"
	report := reduce("abc", density, extraction, nil, nil)
	if report.Resolution != "LOWER_RESOLUTION" || report.Reason != "DENSITY_PATH_UNKNOWN" ||
		report.Coordinates.SourceReceipts != 1 || report.Coordinates.Unknowns != 1 {
		t.Fatalf("report=%#v", report)
	}
}

func TestEvidenceUsesVersionedLowercaseFields(t *testing.T) {
	density, extraction := exactReports()
	report := reduce("abc", density, extraction, []string{"a.go", "c.go"}, nil)
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schema", "decision", "resolution", "coordinates", "indicators", "proofs", "effects"} {
		if _, ok := decoded[key]; !ok {
			t.Fatalf("missing JSON field %q", key)
		}
	}
	if _, ok := decoded["Schema"]; ok {
		t.Fatal("unexpected Go field name in versioned JSON")
	}
}

func TestReduceLowersNonCanonicalPath(t *testing.T) {
	density, extraction := exactReports()
	extraction.Subjects[0].Files[0] = "../a.go"
	report := reduce("abc", density, extraction, nil, nil)
	if report.Resolution != "LOWER_RESOLUTION" || report.Reason != "EXTRACTION_PATH_UNKNOWN" {
		t.Fatalf("report=%#v", report)
	}
}

func TestReduceRejectsKnownExtractionResidual(t *testing.T) {
	density, extraction := exactReports()
	extraction.Unhandled = []string{"pending.go"}
	report := reduce("abc", density, extraction, []string{"a.go", "c.go"}, nil)
	if report.Resolution != "EXACT" || report.Reason != "EXTRACTION_RESIDUAL_PRESENT" ||
		report.Coordinates.Unknowns != 0 {
		t.Fatalf("report=%#v", report)
	}
}
