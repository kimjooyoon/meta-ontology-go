package analyzer

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// ValidateLocalityEnvelope is a read-only, transactional handoff check. It
// rejects missing, relabeled, or tampered locality without changing either IR.
func ValidateLocalityEnvelope(base semantic.IR, result SemanticAdapterResult) error {
	expected, err := LocalityEnvelopeFor(base, result)
	if err != nil {
		return adapterError(AdapterLocalityConfig, "", "", err.Error())
	}
	if result.Locality.Digest != expected.Digest || result.Locality.Canonical() != expected.Canonical() {
		return adapterError(AdapterLocalityConfig, "", "", "locality envelope does not match base closure")
	}
	return nil
}
func validLocalityEnvelope(result SemanticAdapterResult) bool {
	if err := result.Locality.Validate(); err != nil {
		return false
	}
	if result.Locality.BaseDigest != localityBaseDigest(result) {
		return false
	}
	for _, key := range result.Locality.PreservedFacts {
		if !result.IR.Graph.HasFact(key) {
			return false
		}
	}
	touched := localityTouchedFromPreserved(result.Locality.PreservedFacts, result.IR)
	return equalLocalityIDs(touched, result.Locality.Touched) &&
		equalLocalityIDs(localityAffected(result.Locality.PreservedFacts, touched), result.Locality.Affected)
}
func localityTouched(base, result semantic.IR) []semantic.ID {
	touched := make(map[semantic.ID]struct{})
	for _, fact := range result.Graph.DeterministicFacts() {
		if base.Graph.HasFact(fact.Key()) {
			continue
		}
		touched[fact.Subject] = struct{}{}
		touched[fact.Object] = struct{}{}
	}
	return sortedLocalityIDsFromSet(touched)
}
func localityTouchedFromPreserved(preserved []semantic.FactKey, result semantic.IR) []semantic.ID {
	base := make(map[semantic.FactKey]struct{}, len(preserved))
	for _, key := range preserved {
		base[key] = struct{}{}
	}
	touched := make(map[semantic.ID]struct{})
	for _, fact := range result.Graph.DeterministicFacts() {
		if _, ok := base[fact.Key()]; ok {
			continue
		}
		touched[fact.Subject] = struct{}{}
		touched[fact.Object] = struct{}{}
	}
	return sortedLocalityIDsFromSet(touched)
}
