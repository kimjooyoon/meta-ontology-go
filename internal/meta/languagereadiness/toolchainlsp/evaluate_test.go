package toolchainlsp

import (
	"strings"
	"testing"
)

func canonicalConcept() ConceptBinding {
	return ConceptBinding{ArtifactDecision: DecisionPass, ArtifactDigest: "sha256:" + strings.Repeat("1", 64),
		ConceptID: "toolchain-lsp", MetaOperation: MetaOperation, Stage: "OPERATING",
		CodeBindings: append([]string(nil), expectedCodeBindings...), MetricBindings: MetricIDs(), UseCaseBindings: 3}
}

func TestEvaluateExecutesFixedLSPCorpus(t *testing.T) {
	head := strings.Repeat("a", 40)
	report := Evaluate(head, CanonicalCorpus(), canonicalConcept())
	if err := Validate(report, head); err != nil { t.Fatal(err) }
}

func TestUnknownConceptDecisionLowersResolution(t *testing.T) {
	concept := canonicalConcept(); concept.ArtifactDecision = "UNKNOWN"
	report := Evaluate(strings.Repeat("b", 40), CanonicalCorpus(), concept)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionInvariant || report.Reason != "TOOLCHAIN_LSP_DECISION_UNKNOWN" {
		t.Fatalf("report=%#v", report)
	}
}

func TestCorpusDriftFailsClosed(t *testing.T) {
	corpus := CanonicalCorpus(); corpus.Cases = corpus.Cases[:21]
	report := Evaluate(strings.Repeat("c", 40), corpus, canonicalConcept())
	if report.Decision != DecisionFailClosed || report.Summary.CorpusDrift != 1 { t.Fatalf("report=%#v", report) }
}

func TestReportTamperingFailsValidation(t *testing.T) {
	head := strings.Repeat("d", 40)
	report := Evaluate(head, CanonicalCorpus(), canonicalConcept())
	report.Summary.RepositoryWrites = 1
	if err := Validate(report, head); err == nil { t.Fatal("tampered report passed") }
}
