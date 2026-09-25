package main

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func marshalAnalyzeDelta(delta analyzer.SemanticNormalizedDelta, authority, observed semantic.IR) ([]byte, error) {
	if delta.SignatureFacts == nil {
		delta.SignatureFacts = []analyzer.NormalizedSignatureFact{}
	}
	if delta.CandidateFacts == nil {
		delta.CandidateFacts = []analyzer.NormalizedCandidateFact{}
	}
	if delta.DeferredFacts == nil {
		delta.DeferredFacts = []analyzer.NormalizedDeferredFact{}
	}
	if delta.DeferredImplementation == nil {
		delta.DeferredImplementation = []analyzer.ImplementationObservation{}
	}
	if delta.DeferredDetails == nil {
		delta.DeferredDetails = []analyzer.DeferredImplementationDetail{}
	}
	if delta.DeferredSlots == nil {
		delta.DeferredSlots = []analyzer.ProtectedSlotObservation{}
	}
	payload, err := json.Marshal(analyzeDeltaOutput{SemanticNormalizedDelta: delta, AuthoritySemanticDigest: authority.StableHash(), ObservedSemanticDigest: observed.StableHash(), SemanticEqual: true, WriteEffect: analyzer.ReconcileNoWrite})
	if err != nil {
		return nil, err
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return nil, errDiagnosticLimit
	}
	return append(payload, '\n'), nil
}
