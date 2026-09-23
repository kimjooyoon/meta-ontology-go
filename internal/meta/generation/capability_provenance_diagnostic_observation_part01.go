package generation

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

const (
	CapabilityProvenanceDiagnosticInputSchema       = "gooo/lsp-diagnostic-provenance/v1"
	CapabilityProvenanceDiagnosticObservationSchema = "gooo/capability-provenance-diagnostic-observation/v1"
)

type CapabilityProvenanceDiagnosticObservationInput struct {
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
	DiagnosticMapDigest      string
	GeneratedDigest          string
	DeltaSeries              CapabilityProvenanceDeltaSeries
}

type CapabilityProvenanceDiagnosticObservation struct {
	Schema                   string                             `json:"schema"`
	Decision                 string                             `json:"decision"`
	Reason                   string                             `json:"reason"`
	SourceDigest             string                             `json:"source_digest,omitempty"`
	SemanticDigest           string                             `json:"semantic_digest,omitempty"`
	GeneratedDigest          string                             `json:"generated_digest,omitempty"`
	DocumentProvenanceDigest string                             `json:"document_provenance_digest,omitempty"`
	DiagnosticMapDigest      string                             `json:"diagnostic_map_digest,omitempty"`
	ReverseObservationDigest string                             `json:"reverse_observation_digest,omitempty"`
	DeltaSeriesDigest        string                             `json:"delta_series_digest,omitempty"`
	Metrics                  CapabilityProvenanceSurfaceMetrics `json:"metrics"`
	NonAuthorizing           bool                               `json:"non_authorizing"`
	ObservationDigest        string                             `json:"observation_digest"`
}

func ObserveCapabilityProvenanceSurfaceFromDocumentWithDiagnostics(input CapabilityProvenanceDiagnosticObservationInput) CapabilityProvenanceDiagnosticObservation {
	observation := CapabilityProvenanceDiagnosticObservation{
		Schema:                   CapabilityProvenanceDiagnosticObservationSchema,
		Decision:                 CapabilityProvenanceSurfaceUnknown,
		Reason:                   "PROVENANCE_SURFACE_BINDING_UNKNOWN",
		SourceDigest:             input.SourceDigest,
		SemanticDigest:           input.SemanticDigest,
		GeneratedDigest:          input.GeneratedDigest,
		DocumentProvenanceDigest: input.DocumentProvenanceDigest,

		NonAuthorizing: true,
	}
	if input.Schema == "" {
		return finalizeCapabilityProvenanceDiagnosticObservation(observation, CapabilityProvenanceSurfaceUnknown, "MISSING_DIAGNOSTIC_PROVENANCE_SCHEMA")
	}
	if input.Schema != CapabilityProvenanceDiagnosticInputSchema {
		return finalizeCapabilityProvenanceDiagnosticObservation(observation, CapabilityProvenanceSurfaceRefuted, "MALFORMED_DIAGNOSTIC_PROVENANCE_SCHEMA")
	}
	if input.DiagnosticMapDigest == "" {
		return finalizeCapabilityProvenanceDiagnosticObservation(observation, CapabilityProvenanceSurfaceUnknown, "MISSING_DIAGNOSTIC_MAP_DIGEST")
	}
	if !cache.Digest(input.DiagnosticMapDigest).Known() {
		return finalizeCapabilityProvenanceDiagnosticObservation(observation, CapabilityProvenanceSurfaceRefuted, "MALFORMED_DIAGNOSTIC_MAP_DIGEST")
	}
	observation.DiagnosticMapDigest = input.DiagnosticMapDigest
	base := ObserveCapabilityProvenanceSurfaceFromDocument(CapabilityProvenanceDocumentObservationInput{
		Schema:                   CapabilityProvenanceDocumentObservationSchema,
		SubjectDigest:            input.SubjectDigest,
		SourceDigest:             input.SourceDigest,
		SemanticDigest:           input.SemanticDigest,
		ProfileDigest:            input.ProfileDigest,
		ToolchainDigest:          input.ToolchainDigest,
		ContractDigest:           input.ContractDigest,
		SymbolMapDigest:          input.SymbolMapDigest,
		ReferenceMapDigest:       input.ReferenceMapDigest,
		DocumentProvenanceDigest: input.DocumentProvenanceDigest,
		GeneratedDigest:          input.GeneratedDigest,
		DeltaSeries:              input.DeltaSeries,
	})
	observation.Decision = base.Decision
	observation.Reason = base.Reason
	observation.DeltaSeriesDigest = base.DeltaSeriesDigest
	observation.Metrics = base.Metrics
	if base.Decision == CapabilityProvenanceSurfaceClosed {
		observation.ReverseObservationDigest = capabilityProvenanceDocumentDiagnosticReverseDigest(
			input.DocumentProvenanceDigest, input.DiagnosticMapDigest,
		)
		observation.Reason = "DOCUMENT_AND_DIAGNOSTIC_SURFACE_BOUND"
	}
	return finalizeCapabilityProvenanceDiagnosticObservation(observation, observation.Decision, observation.Reason)
}

func capabilityProvenanceDocumentDiagnosticReverseDigest(documentDigest, diagnosticDigest string) string {
	return cache.HashBytes([]byte(strings.Join([]string{
		"gooo/capability-provenance-document-diagnostic-reverse/v1",
		documentDigest,
		diagnosticDigest,
	}, "\x1f"))).String()
}

func finalizeCapabilityProvenanceDiagnosticObservation(observation CapabilityProvenanceDiagnosticObservation, decision, reason string) CapabilityProvenanceDiagnosticObservation {
	observation.Decision = decision
	observation.Reason = reason
	observation.ObservationDigest = cache.HashBytes([]byte(observation.Canonical())).String()
	return observation
}

func (observation CapabilityProvenanceDiagnosticObservation) Canonical() string {
	return strings.Join([]string{
		CapabilityProvenanceDiagnosticObservationSchema,
		observation.Decision,
		observation.Reason,
		observation.SourceDigest,
		observation.SemanticDigest,
		observation.GeneratedDigest,
		observation.DocumentProvenanceDigest,
		observation.DiagnosticMapDigest,
		observation.ReverseObservationDigest,
		observation.DeltaSeriesDigest,
		strconv.Itoa(observation.Metrics.DeltaCount),
		strconv.Itoa(observation.Metrics.ChangeCount),
		strconv.Itoa(observation.Metrics.ClosedDeltas),
		strconv.Itoa(observation.Metrics.UnknownDeltas),
		strconv.Itoa(observation.Metrics.RefutedDeltas),
		observation.Metrics.FinalVerdict,
		strconv.FormatBool(observation.NonAuthorizing),
	}, "\x1f")
}
func (observation CapabilityProvenanceDiagnosticObservation) Validate() error {
	if observation.Schema != CapabilityProvenanceDiagnosticObservationSchema ||
		!observation.NonAuthorizing || observation.Decision == "" || observation.Reason == "" {
		return errors.New("capability provenance diagnostic observation identity is invalid")
	}
	for _, value := range []string{
		observation.SourceDigest,
		observation.SemanticDigest,
		observation.GeneratedDigest,
		observation.DocumentProvenanceDigest,
		observation.DiagnosticMapDigest,
		observation.ReverseObservationDigest,
		observation.DeltaSeriesDigest,
		observation.ObservationDigest,
	} {
		if value != "" && !cache.Digest(value).Known() {
			return errors.New("capability provenance diagnostic observation contains an invalid digest")
		}
	}
	if observation.Decision == CapabilityProvenanceSurfaceClosed &&
		(observation.SourceDigest == "" || observation.SemanticDigest == "" ||
			observation.GeneratedDigest == "" || observation.DocumentProvenanceDigest == "" ||
			observation.DiagnosticMapDigest == "" || observation.ReverseObservationDigest == "" ||
			observation.DeltaSeriesDigest == "") {
		return errors.New("closed capability provenance diagnostic observation is incomplete")
	}
	if observation.ReverseObservationDigest != "" &&
		observation.ReverseObservationDigest != capabilityProvenanceDocumentDiagnosticReverseDigest(
			observation.DocumentProvenanceDigest, observation.DiagnosticMapDigest,
		) {
		return errors.New("capability provenance diagnostic reverse digest mismatch")
	}
	if observation.ObservationDigest != cache.HashBytes([]byte(observation.Canonical())).String() {
		return errors.New("capability provenance diagnostic observation digest mismatch")
	}
	return nil
}
