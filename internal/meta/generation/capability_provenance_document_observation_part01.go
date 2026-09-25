package generation

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/analysisprovenance"
)

// CapabilityProvenanceDocumentObservationSchema identifies the LSP v3 document
// provenance payload consumed by this non-authorizing generation bridge.
const CapabilityProvenanceDocumentObservationSchema = "gooo/lsp-document-provenance/v3"

type CapabilityProvenanceDocumentObservationInput struct {
	Schema                   string
	SubjectDigest            string
	SourceDigest             string
	SemanticDigest           string
	ProfileDigest            string
	ToolchainDigest          string
	ContractDigest           string
	SymbolMapDigest          string
	ReferenceMapDigest       string
	DocumentProvenanceDigest string
	GeneratedDigest          string
	DeltaSeries              CapabilityProvenanceDeltaSeries
}

// ObserveCapabilityProvenanceSurfaceFromDocument proves that a document
// provenance payload is the expected source/IR/editor reverse-observation
// identity before forwarding its provenance digest to the generation surface.
// It never authorizes execution, adoption, promotion, or authentication.
func ObserveCapabilityProvenanceSurfaceFromDocument(input CapabilityProvenanceDocumentObservationInput) CapabilityProvenanceSurfaceObservation {
	base := CapabilityProvenanceSurfaceObservationInput{
		SourceDigest:             input.SourceDigest,
		SemanticDigest:           input.SemanticDigest,
		GeneratedDigest:          input.GeneratedDigest,
		ReverseObservationDigest: input.DocumentProvenanceDigest,
		DeltaSeries:              input.DeltaSeries,
	}
	if input.Schema == "" {
		return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceUnknown, "MISSING_DOCUMENT_PROVENANCE_SCHEMA")
	}
	if input.Schema != CapabilityProvenanceDocumentObservationSchema {
		return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceRefuted, "MALFORMED_DOCUMENT_PROVENANCE_SCHEMA")
	}
	for _, required := range []struct {
		name  string
		value string
	}{
		{name: "SUBJECT_DIGEST", value: input.SubjectDigest},
		{name: "SOURCE_DIGEST", value: input.SourceDigest},
		{name: "SEMANTIC_DIGEST", value: input.SemanticDigest},
		{name: "PROFILE_DIGEST", value: input.ProfileDigest},
		{name: "TOOLCHAIN_DIGEST", value: input.ToolchainDigest},
		{name: "CONTRACT_DIGEST", value: input.ContractDigest},
		{name: "SYMBOL_MAP_DIGEST", value: input.SymbolMapDigest},
		{name: "REFERENCE_MAP_DIGEST", value: input.ReferenceMapDigest},
		{name: "DOCUMENT_PROVENANCE_DIGEST", value: input.DocumentProvenanceDigest},
		{name: "GENERATED_DIGEST", value: input.GeneratedDigest},
	} {
		if required.value == "" {
			return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceUnknown, "MISSING_"+required.name)
		}
		if !cache.Digest(required.value).Known() {
			return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceRefuted, "MALFORMED_"+required.name)
		}
	}
	if input.SubjectDigest != input.SourceDigest {
		return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceRefuted, "DOCUMENT_SUBJECT_MISMATCH")
	}
	expectedDocumentDigest := analysisprovenance.DocumentDigestWithSymbolMapAndReferences(
		input.SourceDigest, input.SemanticDigest, input.ProfileDigest, input.ToolchainDigest,
		input.ContractDigest, input.SymbolMapDigest, input.ReferenceMapDigest,
	)
	if input.DocumentProvenanceDigest != expectedDocumentDigest {
		return finalizeCapabilityProvenanceDocumentObservation(base, CapabilityProvenanceSurfaceRefuted, "DOCUMENT_PROVENANCE_DIGEST_MISMATCH")
	}
	return ObserveCapabilityProvenanceSurface(base)
}

func finalizeCapabilityProvenanceDocumentObservation(input CapabilityProvenanceSurfaceObservationInput, decision, reason string) CapabilityProvenanceSurfaceObservation {
	observation := CapabilityProvenanceSurfaceObservation{
		Schema:                   CapabilityProvenanceSurfaceObservationSchema,
		Decision:                 decision,
		Reason:                   reason,
		SourceDigest:             input.SourceDigest,
		SemanticDigest:           input.SemanticDigest,
		GeneratedDigest:          input.GeneratedDigest,
		ReverseObservationDigest: input.ReverseObservationDigest,
		DeltaSeriesDigest:        input.DeltaSeries.SeriesDigest,
		NonAuthorizing:           true,
	}
	return finalizeCapabilityProvenanceSurfaceObservation(observation)
}

func validateCapabilityProvenanceDocumentObservationInput(input CapabilityProvenanceDocumentObservationInput) error {
	if input.Schema != CapabilityProvenanceDocumentObservationSchema || input.SubjectDigest != input.SourceDigest {
		return errors.New("capability provenance document observation identity is invalid")
	}
	return nil
}
