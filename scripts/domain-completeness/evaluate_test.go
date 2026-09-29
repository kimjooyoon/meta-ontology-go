package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageutility"
)

func TestClassifyKeepsUnknownDistinctFromProgress(t *testing.T) {
	tests := []struct {
		name        string
		numerator   int
		denominator int
		unknown     int
		refuted     bool
		want        string
	}{
		{name: "closed", numerator: 4, denominator: 4, want: "PASS"},
		{name: "measured gap", numerator: 2, denominator: 4, want: "PROGRESS"},
		{name: "unobserved", denominator: 4, want: "UNKNOWN"},
		{name: "unknown frontier", numerator: 2, denominator: 4, unknown: 1, want: "UNKNOWN"},
		{name: "contradiction", numerator: 4, denominator: 4, refuted: true, want: "FAIL_CLOSED"},
		{name: "invalid population", numerator: 5, denominator: 4, want: "FAIL_CLOSED"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := classify(test.numerator, test.denominator, test.unknown, test.refuted); got != test.want {
				t.Fatalf("classify() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestDecisionKeepsUnknownAndRefutedEvidenceDistinct(t *testing.T) {
	progress := Dimension{ID: "generation_coverage", Status: "PROGRESS",
		FirstUnresolved: &Frontier{NextOperation: "COLLECT_GENERATION"}}
	unknown := Dimension{ID: "provenance_integrity", Status: "UNKNOWN",
		FirstUnresolved: &Frontier{NextOperation: "COLLECT_PROVENANCE"}}
	refuted := Dimension{ID: "boundary_coverage", Status: "FAIL_CLOSED",
		FirstUnresolved: &Frontier{NextOperation: "REPAIR_BOUNDARY"}}

	decision, _, next, _ := decide([]Dimension{progress, unknown}, nil, "PASS")
	if decision != "UNKNOWN" || next != "COLLECT_PROVENANCE" {
		t.Fatalf("unknown frontier decision = %q, next %q", decision, next)
	}
	decision, _, next, _ = decide([]Dimension{progress, refuted}, nil, "PASS")
	if decision != "FAIL_CLOSED" || next != "REPAIR_BOUNDARY" {
		t.Fatalf("refuted frontier decision = %q, next %q", decision, next)
	}
}

func TestEvidenceReferencesRequireExactDigestAndSafePath(t *testing.T) {
	root := t.TempDir()
	payload := []byte("{\"decision\":\"PASS\"}")
	if err := os.MkdirAll(filepath.Join(root, "evidence"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "evidence", "case.json"), payload, 0o644); err != nil {
		t.Fatal(err)
	}
	report := languageutility.Report{Cells: []languageutility.CellResult{{
		EvidencePath: "evidence/case.json", EvidenceDigest: digestBytes(payload),
	}}}
	refs, state, _ := validateEvidenceReferences(root, report)
	if state != "PASS" || len(refs) != 1 || refs[0].Digest != digestBytes(payload) {
		t.Fatalf("exact reference result = %#v, %q", refs, state)
	}

	report.Cells[0].EvidenceDigest = "sha256:00"
	_, state, reason := validateEvidenceReferences(root, report)
	if state != "FAIL_CLOSED" || reason != "CELL_EVIDENCE_DIGEST_MISMATCH" {
		t.Fatalf("mismatched digest result = %q / %q", state, reason)
	}

	report.Cells[0].EvidencePath = "../outside.json"
	_, state, reason = validateEvidenceReferences(root, report)
	if state != "FAIL_CLOSED" || reason != "EVIDENCE_PATH_ESCAPES_BUNDLE" {
		t.Fatalf("escaping path result = %q / %q", state, reason)
	}
}

func TestReportDigestIsDeterministic(t *testing.T) {
	report := Report{
		Schema: ReceiptSchema, ProfileID: ProfileID, Decision: "UNKNOWN",
		Dimensions: []Dimension{{ID: "use_case_coverage", Status: "UNKNOWN"}},
		Summary:    Summary{DimensionsTotal: 1, Unknown: 1},
	}
	first, err := reportDigest(report)
	if err != nil {
		t.Fatal(err)
	}
	second, err := reportDigest(report)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("report digests differ: %s and %s", first, second)
	}
}
