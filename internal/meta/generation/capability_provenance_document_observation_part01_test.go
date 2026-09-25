package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/analysisprovenance"
)

func TestCapabilityProvenanceDocumentObservationBridgesReferenceMap(t *testing.T) {
	before := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	after := cloneCapabilityLineageForComposition(before)
	after.Steps[0].Digest = "source-digest-next"
	after.ChainDigest = after.StableHash()
	delta, err := CompareCapabilityProvenanceLineage(before, after)
	if err != nil {
		t.Fatalf("compare delta: %v", err)
	}
	series, err := FoldCapabilityProvenanceDeltas([]CapabilityProvenanceDelta{delta})
	if err != nil {
		t.Fatalf("fold series: %v", err)
	}
	source := cache.HashBytes([]byte("gooo-source")).String()
	semantic := cache.HashBytes([]byte("gooo-semantic-ir")).String()
	profile := cache.HashBytes([]byte("entity-fields-profile")).String()
	toolchain := cache.HashBytes([]byte("go1.27-linux-amd64")).String()
	contract := cache.HashBytes([]byte("gooo-lsp-document-provenance-v3")).String()
	symbols := cache.HashBytes([]byte("symbols")).String()
	references := cache.HashBytes([]byte("references")).String()
	documentDigest := analysisprovenance.DocumentDigestWithSymbolMapAndReferences(source, semantic, profile, toolchain, contract, symbols, references)
	observation := ObserveCapabilityProvenanceSurfaceFromDocument(CapabilityProvenanceDocumentObservationInput{
		Schema:                   CapabilityProvenanceDocumentObservationSchema,
		SubjectDigest:            source,
		SourceDigest:             source,
		SemanticDigest:           semantic,
		ProfileDigest:            profile,
		ToolchainDigest:          toolchain,
		ContractDigest:           contract,
		SymbolMapDigest:          symbols,
		ReferenceMapDigest:       references,
		DocumentProvenanceDigest: documentDigest,
		GeneratedDigest:          cache.HashBytes([]byte("generated-artifact")).String(),
		DeltaSeries:              series,
	})
	if observation.Decision != CapabilityProvenanceSurfaceClosed ||
		observation.Reason != "PROVENANCE_SURFACE_BOUND" ||
		observation.ReverseObservationDigest != documentDigest || !observation.NonAuthorizing {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate observation: %v", err)
	}
}

func TestCapabilityProvenanceDocumentObservationPreservesUnknownAndRefuted(t *testing.T) {
	missing := ObserveCapabilityProvenanceSurfaceFromDocument(CapabilityProvenanceDocumentObservationInput{})
	if missing.Decision != CapabilityProvenanceSurfaceUnknown || missing.Reason != "MISSING_DOCUMENT_PROVENANCE_SCHEMA" {
		t.Fatalf("missing observation = %#v", missing)
	}
	if err := missing.Validate(); err != nil {
		t.Fatalf("validate missing observation: %v", err)
	}
	valid := cache.HashBytes([]byte("bridge-valid")).String()
	wrongDocumentDigest := cache.HashBytes([]byte("bridge-wrong-document")).String()
	refuted := ObserveCapabilityProvenanceSurfaceFromDocument(CapabilityProvenanceDocumentObservationInput{
		Schema:                   CapabilityProvenanceDocumentObservationSchema,
		SubjectDigest:            valid,
		SourceDigest:             valid,
		SemanticDigest:           valid,
		ProfileDigest:            valid,
		ToolchainDigest:          valid,
		ContractDigest:           valid,
		SymbolMapDigest:          valid,
		ReferenceMapDigest:       valid,
		DocumentProvenanceDigest: wrongDocumentDigest,
		GeneratedDigest:          valid,
	})
	if refuted.Decision != CapabilityProvenanceSurfaceRefuted || refuted.Reason != "DOCUMENT_PROVENANCE_DIGEST_MISMATCH" {
		t.Fatalf("refuted observation = %#v", refuted)
	}
	if err := refuted.Validate(); err != nil {
		t.Fatalf("validate refuted observation: %v", err)
	}
}
