package languageconcept

import (
	"os"
	"reflect"
	"testing"
)

func TestCatalogBindsConceptsToMetaCode(t *testing.T) {
	report := Evaluate(os.DirFS("../../.."), Catalog())
	if report.Decision != "PASS" || report.Reason != "LANGUAGE_CONCEPT_CATALOG_BOUND" {
		t.Fatalf("got %s/%s: %v", report.Decision, report.Reason, report.MissingBindings)
	}
	if report.Summary.Concepts != 9 || report.Summary.Unbound != 0 || report.Summary.UnverifiedNovelty != 0 {
		t.Fatalf("summary = %#v", report.Summary)
	}
	if len(report.Indicators) != 7 || len(report.Proofs) != 3 || report.ReportDigest == "" {
		t.Fatalf("incomplete evidence: %#v", report)
	}
}

func TestVerifiedTransactionBindsMetaCodeAndMetrics(t *testing.T) {
	item := Catalog()[8]
	wantCode := []string{"internal/meta/languagereadiness/artifact/predecessorselection", "internal/meta/languagereadiness/artifact/predecessorbinding", "cmd/language-readiness-witness/predecessor-selection"}
	wantMetrics := []string{"gooo.metric.language.predecessor-dynamic-binding-bps.v1", "gooo.metric.language.predecessor-dynamic-coordinates.v1", "gooo.metric.language.predecessor-static-coordinates.guardrail.v1", "gooo.metric.language.predecessor-unknown-coordinates.guardrail.v1", "gooo.metric.language.predecessor-observer-writes.guardrail.v1"}
	if item.ID != "verified-transformation-transaction" || item.MetaOperation != "verify-transformation-transaction" || item.Stage != "OPERATING" {
		t.Fatalf("concept = %#v", item)
	}
	if !reflect.DeepEqual(item.CodeBindings, wantCode) || !reflect.DeepEqual(item.MetricBindings, wantMetrics) {
		t.Fatalf("code=%v metrics=%v", item.CodeBindings, item.MetricBindings)
	}
	if len(item.UseCases) != 1 || item.UseCases[0].ExpectedOutcome != "IMPROVED_STATIC_8_TO_0_DYNAMIC_0_TO_8_BPS_0_TO_10000" {
		t.Fatalf("use cases = %#v", item.UseCases)
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
