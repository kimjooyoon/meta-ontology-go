package analyzer

func normalizedDeltaBindingsMatch(result SemanticAdapterResult) bool {
	var binding *DeltaBinding
	accept := func(candidate DeltaBinding) bool {
		if !candidate.complete() || candidate.SourceDigest != result.SourceDigest ||
			candidate.PolicyDigest != result.PolicyDigest || candidate.ToolchainDigest != result.ToolchainDigest ||
			candidate.RegistryDigest != result.RegistryDigest {
			return false
		}
		if binding == nil {
			copyOf := candidate
			binding = &copyOf
			return true
		}
		return *binding == candidate
	}
	for _, fact := range result.NormalizedDelta.SignatureFacts {
		if !accept(fact.Binding) {
			return false
		}
	}
	for _, candidate := range result.NormalizedDelta.CandidateFacts {
		if !accept(candidate.Binding) {
			return false
		}
	}
	for _, fact := range result.NormalizedDelta.DeferredFacts {
		if !accept(fact.Binding) {
			return false
		}
	}
	for _, observation := range result.NormalizedDelta.DeferredImplementation {
		if !validDigest(observation.SourceDigest) || observation.SourceDigest != result.SourceDigest ||
			observation.PolicyDigest != result.PolicyDigest || observation.ToolchainDigest != result.ToolchainDigest ||
			!validDigest(observation.RegistryDigest) || observation.RegistryDigest != result.RegistryDigest {
			return false
		}
		if binding == nil {
			binding = &DeltaBinding{
				SourceDigest: observation.SourceDigest, BaseDigest: observation.BaseDigest,
				PolicyDigest: observation.PolicyDigest, ToolchainDigest: observation.ToolchainDigest,
				RegistryDigest: observation.RegistryDigest,
			}
		} else if binding.BaseDigest != observation.BaseDigest {
			return false
		}
	}
	for _, detail := range result.NormalizedDelta.DeferredDetails {
		if !accept(detail.Binding) {
			return false
		}
	}
	for _, slot := range result.NormalizedDelta.DeferredSlots {
		if !accept(DeltaBinding{
			SourceDigest: slot.SourceDigest, BaseDigest: slot.BaseDigest,
			PolicyDigest: slot.PolicyDigest, ToolchainDigest: slot.ToolchainDigest,
			RegistryDigest: slot.RegistryDigest,
		}) {
			return false
		}
	}
	return binding != nil
}
