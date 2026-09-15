package analyzer

func normalizedDeltaValid(result SemanticAdapterResult) bool {
	if result.NormalizedDelta.Digest == "" || result.NormalizedDelta.Digest != result.NormalizedDelta.StableHash() {
		return false
	}
	if !validDigest(result.BindingDigest) || result.BindingDigest != semanticAdapterBindingDigest(result) {
		return false
	}
	if result.SlotObservationDigest != protectedSlotObservationDigest(result.SlotObservations) ||
		result.SlotObservationDigest != protectedSlotObservationDigest(result.NormalizedDelta.DeferredSlots) ||
		len(result.SlotObservations) != len(result.NormalizedDelta.DeferredSlots) {
		return false
	}
	if result.ImplementationObservationDigest != implementationObservationDigest(
		result.ImplementationObservations, result.SlotObservations,
	) {
		return false
	}
	if !validLocalityEnvelope(result) {
		return false
	}
	if !implementationObservationsMatch(
		result.ImplementationObservations, result.NormalizedDelta.DeferredImplementation,
	) {
		return false
	}
	if !deferredImplementationDetailsMatch(result) {
		return false
	}
	if !candidateObservationsMatch(result) {
		return false
	}
	if !deferredFactsMatch(result) {
		return false
	}
	if !validDigest(result.RegistryDigest) {
		return false
	}
	if err := validateDeltaShape(result.NormalizedDelta); err != nil {
		return false
	}
	return normalizedDeltaMembersValid(result)
}
