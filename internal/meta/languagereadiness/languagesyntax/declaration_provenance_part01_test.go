package languagesyntax

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/languagesyntax/replay"
	"testing"
)

func TestLanguageSyntaxDeclarationProvenanceBindsGoooInventory(t *testing.T) {
	source := Source{
		RegistryDigest:   digestBytes([]byte("registry")),
		CorpusDigest:     digestBytes([]byte("corpus")),
		GoooFiles:        []replay.FileObservation{{Path: "examples/alpha.gooo"}, {Path: "examples/beta.gooo"}},
		ObservationKnown: true,
		ConceptBound:     true,
	}
	observation := ObserveLanguageSyntaxDeclarationProvenance(source)
	if observation.Decision != LanguageSyntaxDeclarationProvenanceClosed ||
		observation.Reason != "GOOO_DECLARATION_REGISTRY_BOUND" {
		t.Fatalf("observation = %#v", observation)
	}
	if observation.RegisteredCount != 2 ||
		observation.GoooFilesDigest == "" ||
		observation.MissingRegisteredDigest == "" ||
		observation.UnregisteredGoooDigest == "" {
		t.Fatalf("inventory evidence = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestLanguageSyntaxDeclarationProvenancePreservesDriftAndUnknown(t *testing.T) {
	unknown := ObserveLanguageSyntaxDeclarationProvenance(Source{})
	if unknown.Decision != LanguageSyntaxDeclarationProvenanceUnknown ||
		unknown.Reason != "MISSING_DECLARATION_REGISTRY_DIGEST" {
		t.Fatalf("unknown = %#v", unknown)
	}
	if err := unknown.Validate(); err != nil {
		t.Fatalf("unknown Validate() error = %v", err)
	}

	refuted := ObserveLanguageSyntaxDeclarationProvenance(Source{
		RegistryDigest:   digestBytes([]byte("registry")),
		GoooFiles:        []replay.FileObservation{{Path: "examples/alpha.gooo"}},
		UnregisteredGooo: []string{"examples/rogue.gooo"},
	})
	if refuted.Decision != LanguageSyntaxDeclarationProvenanceRefuted ||
		refuted.Reason != "GOOO_DECLARATION_REGISTRY_DRIFT" {
		t.Fatalf("refuted = %#v", refuted)
	}
	if err := refuted.Validate(); err != nil {
		t.Fatalf("refuted Validate() error = %v", err)
	}
}
