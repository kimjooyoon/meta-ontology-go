package languageconcept

import (
	"os"
	"testing"
)

func TestCatalogBindsConceptsToMetaCode(t *testing.T) {
	report := Evaluate(os.DirFS("../../.."), Catalog())
	if report.Decision != "PASS" || report.Reason != "LANGUAGE_CONCEPT_CATALOG_BOUND" {
		t.Fatalf("got %s/%s: %v", report.Decision, report.Reason, report.MissingBindings)
	}
	if report.Summary.Concepts != 8 || report.Summary.Unbound != 0 || report.Summary.UnverifiedNovelty != 0 {
		t.Fatalf("summary = %#v", report.Summary)
	}
	if len(report.Indicators) != 7 || len(report.Proofs) != 3 || report.ReportDigest == "" {
		t.Fatalf("incomplete evidence: %#v", report)
	}
}

func TestMissingMetaCodeFailsClosed(t *testing.T) {
	concepts := Catalog()
	concepts[0].CodeBindings = []string{"missing/meta-program"}
	report := Evaluate(os.DirFS("../../.."), concepts)
	if report.Decision != "FAIL_CLOSED" || report.Summary.Unbound != 1 {
		t.Fatalf("report = %#v", report)
	}
}

func TestUnverifiedNoveltyFailsClosed(t *testing.T) {
	concepts := Catalog()
	concepts[0].NoveltyClaim = true
	report := Evaluate(os.DirFS("../../.."), concepts)
	if report.Decision != "FAIL_CLOSED" || report.Summary.UnverifiedNovelty != 1 {
		t.Fatalf("report = %#v", report)
	}
}
