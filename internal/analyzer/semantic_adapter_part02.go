package analyzer

// AdaptSemantic performs a transactional, explicit mapping. The input IR is
// never mutated, including when a later observation fails preflight.
func AdaptSemantic(input SemanticAdapterInput) (SemanticAdapterResult, error) {
	if err := input.Policy.Validate(); err != nil {
		return SemanticAdapterResult{}, err
	}
	base, err := input.Base.Normalized()
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	baseForLocality, err := base.Normalized()
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	baseDigest := base.StableHash()
	if err := validateObservations(input.Analysis); err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := validateSlotObservations(input.SlotObservations, input.SourceDigest, baseDigest,
		input.Policy.Digest(), input.ToolchainDigest, input.Registry.Digest()); err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction := SemanticAdapterResult{
		IR: base, SourceDigest: input.SourceDigest, PolicyDigest: input.Policy.Digest(),
		ToolchainDigest:       input.ToolchainDigest,
		RegistryDigest:        input.Registry.Digest(),
		SlotObservations:      append([]ProtectedSlotObservation(nil), input.SlotObservations...),
		DeferredCandidates:    copyCandidates(input.Analysis.Delta.Candidates),
		ImplementationDetails: copyDetails(input.Analysis.Delta.ImplementationDetails),
	}
	transaction.SlotObservationDigest = protectedSlotObservationDigest(transaction.SlotObservations)
	transaction.ImplementationObservations = collectImplementationObservations(input.Analysis, base, input)
	transaction.ImplementationObservationDigest = implementationObservationDigest(
		transaction.ImplementationObservations, transaction.SlotObservations,
	)
	if err := addRegisteredNodes(&transaction.IR, input.Analysis.Registrations); err != nil {
		return SemanticAdapterResult{}, err
	}
	if hasMappedObservation(input.Analysis, input.Policy) {
		if err := validateEvidenceConfig(input); err != nil {
			return SemanticAdapterResult{}, err
		}
	}
	if err := adaptFacts(&transaction, input); err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := adaptCandidates(&transaction, input); err != nil {
		return SemanticAdapterResult{}, err
	}
	if err := transaction.IR.Validate(); err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction.NormalizedDelta, err = newSemanticNormalizedDelta(input, baseDigest, transaction)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction.Locality, err = LocalityEnvelopeFor(baseForLocality, transaction)
	if err != nil {
		return SemanticAdapterResult{}, err
	}
	transaction.BindingDigest = semanticAdapterBindingDigest(transaction)
	return transaction, nil
}
