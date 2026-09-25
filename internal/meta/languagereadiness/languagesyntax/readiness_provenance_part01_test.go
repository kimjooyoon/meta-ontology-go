package languagesyntax

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"strings"
	"testing"
)

func TestLanguageSyntaxReadinessProvenanceComposesDeclarations(t *testing.T) {
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
			RegistryDigest:   digestBytes([]byte("registry")),
			CorpusDigest:     digestBytes([]byte("corpus")),
			GoooFiles:        []replay.FileObservation{{Path: "examples/alpha.gooo"}},
			ObservationKnown: true,
			ConceptBound:     true,
		},
		Summary: Summary{Satisfied: totalCases, Total: totalCases},
	})
	observation := ObserveLanguageSyntaxReadinessProvenance(report)
	if observation.Decision != LanguageSyntaxReadinessProvenanceClosed ||
		observation.Reason != "LANGUAGE_SYNTAX_READINESS_EVIDENCE_COMPOSED" {
		t.Fatalf("observation = %#v", observation)
	}
	if observation.ReverseObservationDigest == "" {
		t.Fatal("missing composed reverse observation digest")
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLanguageSyntaxReadinessProvenancePreservesUnknownAndRefuted(t *testing.T) {
	unknown := ObserveLanguageSyntaxReadinessProvenance(Report{})
	if unknown.Decision != LanguageSyntaxReadinessProvenanceUnknown ||
		unknown.Reason != "LANGUAGE_SYNTAX_READINESS_EVIDENCE_UNKNOWN" {
		t.Fatalf("unknown = %#v", unknown)
	}
	if err := unknown.Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}

	head := strings.Repeat("b", 40)
	refuted := seal(Report{
		Schema:  ReportSchema,
		HeadSHA: head,
		Source: Source{
			RegistryDigest:   digestBytes([]byte("registry")),
			GoooFiles:        []replay.FileObservation{{Path: "examples/alpha.gooo"}},
			UnregisteredGooo: []string{"examples/rogue.gooo"},
		},
	})
	observation := ObserveLanguageSyntaxReadinessProvenance(refuted)
	if observation.Decision != LanguageSyntaxReadinessProvenanceRefuted ||
		observation.Reason != "LANGUAGE_SYNTAX_READINESS_EVIDENCE_REFUTED" {
		t.Fatalf("refuted = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("refuted Validate() error = %v", err)
	}
}
