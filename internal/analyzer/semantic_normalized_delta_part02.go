package analyzer

import (
	"strings"
)

func (f NormalizedDeferredFact) canonical() string {
	var builder strings.Builder
	builder.WriteString("deferred-fact\n")
	builder.WriteString(f.Binding.canonical())
	builder.WriteString(sourceFactCanonical(f.Fact))
	return builder.String()
}
func (f NormalizedCandidateFact) canonical() string {
	var builder strings.Builder
	builder.WriteString("candidate\n")
	builder.WriteString(f.Binding.canonical())
	writeBindingField(&builder, f.ObservationDigest)
	writeBindingField(&builder, string(f.SourceRelation))
	writeBindingField(&builder, string(f.Origin))
	writeBindingField(&builder, f.Subject.String())
	for _, option := range f.Options {
		writeBindingField(&builder, option.String())
	}
	for _, fact := range f.Facts {
		builder.WriteString(fact.Canonical())
	}
	for _, evidence := range f.Evidence {
		builder.WriteString(evidence.Canonical())
	}
	writeSemanticSpan(&builder, f.Span)
	writeBindingField(&builder, f.Reason)
	return builder.String()
}

// SemanticNormalizedDelta is the machine-readable handoff boundary. The
// three slices deliberately cannot be confused by a status bit or a string.
type SemanticNormalizedDelta struct {
	SchemaVersion          string                         `json:"schema_version"`
	SignatureFacts         []NormalizedSignatureFact      `json:"signature_facts"`
	CandidateFacts         []NormalizedCandidateFact      `json:"candidate_facts"`
	DeferredFacts          []NormalizedDeferredFact       `json:"deferred_facts"`
	DeferredImplementation []ImplementationObservation    `json:"deferred_implementation"`
	DeferredDetails        []DeferredImplementationDetail `json:"deferred_details"`
	DeferredSlots          []ProtectedSlotObservation     `json:"deferred_slots"`
	Digest                 string                         `json:"digest"`
}
