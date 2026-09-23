package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/analysisprovenance"
)

func TestCapabilityProvenanceDiagnosticObservationComposesDocumentAndDiagnostic(t *testing.T) {
	before := testCapabilityLifecycleLineage(CapabilityLineageComplete)
	after := cloneCapabilityLineageForComposition(before)
	after.Steps[0].Digest = "source-digest-diagnostic"
	after.ChainDigest = after.StableHash()
	delta, err := CompareCapabilityProvenanceLineage(before, after)
	if err != nil {
		t.Fatalf("compare delta: %v", err)
	}
	series, err := FoldCapabilityProvenanceDeltas([]CapabilityProvenanceDelta{delta})
	if err != nil {
		t.Fatalf("fold series: %v", err)
	}
	source := cache.HashBytes([]byte("diagnostic-source")).String()
	semantic := cache.HashBytes([]byte("diagnostic-semantic-ir")).String()
	profile := cache.HashBytes([]byte("diagnostic-profile")).String()
	toolchain := cache.HashBytes([]byte("diagnostic-toolchain")).String()
	contract := cache.HashBytes([]byte("diagnostic-contract")).String()
	symbols := cache.HashBytes([]byte("diagnostic-symbols")).String()
	references := cache.HashBytes([]byte("diagnostic-references")).String()
	documentDigest := analysisprovenance.DocumentDigestWithSymbolMapAndReferences(
		source, semantic, profile, toolchain, contract, symbols, references,
	)
	diagnosticMap := cache.HashBytes([]byte("diagnostic-map")).String()
	observation := ObserveCapabilityProvenanceSurfaceFromDocumentWithDiagnostics(CapabilityProvenanceDiagnosticObservationInput{
		Schema:                   CapabilityProvenanceDiagnosticInputSchema,
		SubjectDigest:            source,
		SourceDigest:             source,
		SemanticDigest:           semantic,
		ProfileDigest:            profile,
		ToolchainDigest:          toolchain,
		ContractDigest:           contract,
		SymbolMapDigest:          symbols,
		ReferenceMapDigest:       references,
		DocumentProvenanceDigest: documentDigest,
		DiagnosticMapDigest:      diagnosticMap,
		GeneratedDigest:          cache.HashBytes([]byte("diagnostic-generated")).String(),
		DeltaSeries:              series,
	})
	if observation.Decision != CapabilityProvenanceSurfaceClosed ||
		observation.Reason != "DOCUMENT_AND_DIAGNOSTIC_SURFACE_BOUND" ||
		observation.ReverseObservationDigest == "" || observation.Metrics.DeltaCount != 1 {
		t.Fatalf("observation = %#v", observation)
	}
	if err := observation.Validate(); err != nil {
		t.Fatalf("validate observation: %v", err)
	}
}

func TestCapabilityProvenanceDiagnosticObservationPreservesUnknownAndRefuted(t *testing.T) {
	missing := ObserveCapabilityProvenanceSurfaceFromDocumentWithDiagnostics(CapabilityProvenanceDiagnosticObservationInput{})
	if missing.Decision != CapabilityProvenanceSurfaceUnknown ||
		missing.Reason != "MISSING_DIAGNOSTIC_PROVENANCE_SCHEMA" {
		t.Fatalf("missing observation = %#v", missing)
	}
	if err := missing.Validate(); err != nil {
		t.Fatalf("validate missing observation: %v", err)
	}
	valid := cache.HashBytes([]byte("diagnostic-valid")).String()
	refuted := ObserveCapabilityProvenanceSurfaceFromDocumentWithDiagnostics(CapabilityProvenanceDiagnosticObservationInput{
		Schema:              CapabilityProvenanceDiagnosticInputSchema,
		DiagnosticMapDigest: "not-a-digest",
		SourceDigest:        valid,
	})
	if refuted.Decision != CapabilityProvenanceSurfaceRefuted ||
		refuted.Reason != "MALFORMED_DIAGNOSTIC_MAP_DIGEST" {
		t.Fatalf("refuted observation = %#v", refuted)
	}
	if err := refuted.Validate(); err != nil {
		t.Fatalf("validate refuted observation: %v", err)
	}
}
