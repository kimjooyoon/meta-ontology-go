package analyzer

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func validateDeltaShape(delta SemanticNormalizedDelta) error {
	if delta.SchemaVersion != semanticNormalizedDeltaSchema {
		return fmt.Errorf("normalized delta schema is %q", delta.SchemaVersion)
	}
	for _, fact := range delta.SignatureFacts {
		if !knownAnalyzerRelation(fact.SourceRelation) || fact.Fact.Status != semantic.FactDeterministic ||
			!fact.Fact.Predicate.Valid() || fact.Evidence.Status != semantic.FactDeterministic {
			return fmt.Errorf("signature fact binding or status is incomplete")
		}
	}
	for _, candidate := range delta.CandidateFacts {
		if !candidate.Subject.Valid() || !knownAnalyzerRelation(candidate.SourceRelation) {
			return fmt.Errorf("candidate binding or identity is incomplete")
		}
		for _, option := range candidate.Options {
			if _, err := semantic.ParseIdentity(option.String()); err != nil {
				return err
			}
		}
		for _, fact := range candidate.Facts {
			if fact.Status != semantic.FactCandidate || !fact.Predicate.Valid() {
				return fmt.Errorf("candidate semantic fact is not typed")
			}
		}
	}
	for _, fact := range delta.DeferredFacts {
		if !validSourceFact(fact.Fact) {
			return fmt.Errorf("deferred fact is invalid")
		}
	}
	for _, observation := range delta.DeferredImplementation {
		if !validImplementationObservation(observation) {
			return fmt.Errorf("deferred implementation observation is invalid")
		}
	}
	for _, detail := range delta.DeferredDetails {
		if !validateDeferredImplementationDetail(detail) {
			return fmt.Errorf("deferred implementation detail is incomplete")
		}
	}
	for _, slot := range delta.DeferredSlots {
		if !validProtectedSlotObservation(slot) {
			return fmt.Errorf("deferred protected slot is incomplete")
		}
	}
	return nil
}

func writeSemanticSpan(builder *strings.Builder, span semantic.Span) {
	writeBindingField(builder, span.File)
	writeBindingField(builder, intString(span.Start.Offset))
	writeBindingField(builder, intString(span.Start.Line))
	writeBindingField(builder, intString(span.Start.Column))
	writeBindingField(builder, intString(span.End.Offset))
	writeBindingField(builder, intString(span.End.Line))
	writeBindingField(builder, intString(span.End.Column))
}
