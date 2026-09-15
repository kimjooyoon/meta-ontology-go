package analyzer

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func localityAffected(baseFacts []semantic.FactKey, touched []semantic.ID) []semantic.ID {
	closure := make(map[semantic.ID]struct{}, len(touched))
	for _, id := range touched {
		closure[id] = struct{}{}
	}
	for _, fact := range baseFacts {
		_, subjectTouched := closure[fact.Subject]
		_, objectTouched := closure[fact.Object]
		if subjectTouched || objectTouched {
			closure[fact.Subject] = struct{}{}
			closure[fact.Object] = struct{}{}
		}
	}
	return sortedLocalityIDsFromSet(closure)
}
func localityFactKeys(facts []semantic.Fact) []semantic.FactKey {
	keys := make([]semantic.FactKey, 0, len(facts))
	for _, fact := range facts {
		keys = append(keys, fact.Key())
	}
	return sortedLocalityFacts(keys)
}
func localityBaseDigest(result SemanticAdapterResult) string {
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		return fact.Binding.BaseDigest
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		return candidate.Binding.BaseDigest
	}
	for _, fact := range result.NormalizedDelta.DeferredFacts {
		return fact.Binding.BaseDigest
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		return observation.BaseDigest
	}
	for _, detail := range result.NormalizedDelta.DeferredDetails {
		return detail.Binding.BaseDigest
	}
	for _, slot := range result.NormalizedDelta.DeferredSlots {
		return slot.BaseDigest
	}
	return ""
}
func validateLocalityIDs(ids []semantic.ID) error {
	seen := make(map[semantic.ID]struct{}, len(ids))
	for _, id := range ids {
		if _, err := semantic.ParseIdentity(id.String()); err != nil {
			return err
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate locality ID: %s", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}
