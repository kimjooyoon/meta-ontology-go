package languagesyntax

import (
	"strings"
	"testing"
)

func TestLanguageSyntaxProvenanceBindsReportEvidence(t *testing.T) {
	head := strings.Repeat("a", 40)
	report := seal(Report{
		Schema:        ReportSchema,
		Decision:      DecisionPass,
		Reason:        "LANGUAGE_SYNTAX_ROUNDTRIP_PROVEN",
		Resolution:    ResolutionExact,
		Producer:      "languagesyntax.Evaluate",
		Consumer:      "self-improvement-cycle",
		MetaOperation: "prove-language-syntax-roundtrip",
		HeadSHA:       head,
		Source: Source{
			ExpectedHeadSHA:  head,
			RegistryDigest:   digestBytes([]byte("registry")),
			CorpusDigest:     digestBytes([]byte("corpus")),
			ObservationKnown: true,
			ConceptBound:     true,
		},
		Summary: Summary{Satisfied: totalCases, Total: totalCases},
	})
	observation := ObserveLanguageSyntaxProvenance(report)
	if observation.Decision != LanguageSyntaxProvenanceClosed {
		t.Fatalf("decision = %q, want %q", observation.Decision, LanguageSyntaxProvenanceClosed)
	}
	if observation.Reason != "LANGUAGE_SYNTAX_EVIDENCE_BOUND" {
		t.Fatalf("reason = %q", observation.Reason)
	}
	if observation.ReportedDecision != DecisionPass {
		t.Fatalf("reported decision = %q", observation.ReportedDecision)
	}
	if observation.SourceDigest == "" || observation.CaseEvidenceDigest == "" ||
		observation.IndicatorDigest == "" || observation.ReverseObservationDigest == "" {
		t.Fatal("closed observation did not bind all evidence digests")
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLanguageSyntaxProvenancePreservesUnknownAndRefuted(t *testing.T) {
	missing := ObserveLanguageSyntaxProvenance(Report{})
	if missing.Decision != LanguageSyntaxProvenanceUnknown ||
		missing.Reason != "MISSING_REPORT_SCHEMA" {
		t.Fatalf("missing report = %#v", missing)
	}
	if err := missing.Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}

	head := strings.Repeat("b", 40)
	report := seal(Report{
		Schema:  ReportSchema,
		HeadSHA: head,
		Source:  Source{ExpectedHeadSHA: head},
	})
	report.ReportDigest = digestBytes([]byte("tampered"))
	refuted := ObserveLanguageSyntaxProvenance(report)
	if refuted.Decision != LanguageSyntaxProvenanceRefuted ||
		refuted.Reason != "REPORT_DIGEST_MISMATCH" {
		t.Fatalf("refuted report = %#v", refuted)
	}
	if err := refuted.Validate(); err != nil {
		t.Fatalf("refuted Validate() error = %v", err)
	}
}
