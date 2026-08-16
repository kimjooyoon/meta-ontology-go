package coupling

type oracleValidation struct {
	decision Decision
	reason   Reason
}

type normalizedSemantic struct {
	digest string
	facts  []string
}

type registryView struct {
	bySurface map[string]CodeBinding
	bySymbol  map[string]CodeBinding
	digest    string
}

type receiptView struct {
	bySurface map[string]CouplingReceipt
	valid     []string
}

type pathView struct {
	digest   string
	counts   ObservationCounts
	decision Decision
	reason   Reason
}

func Evaluate(input Input) Output {
	output := Output{Schema: SchemaV1, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input)}
	output.ObservationCounts = inputObservationCounts(input)
	if !resourceBindingsEqual(input.Config.ResourceBinding, input.ResourceRegistry) {
		return finish(output, DecisionUnknown, ReasonResourceUnbound)
	}

	if issue := validateRequiredInput(input); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	resources, issue := normalizeResources(input.ResourceReceipts, input.Config.ResourceBinding)
	output.Resources = resources
	output.ObservationCounts.ResourceReceipts = uint64(len(input.ResourceReceipts))
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	before, err := normalizeSemantic(input.SemanticBefore)
	if err != nil {
		return finish(output, DecisionFailClosed, ReasonPathMalformed)
	}
	after, err := normalizeSemantic(input.SemanticAfter)
	if err != nil {
		return finish(output, DecisionFailClosed, ReasonPathMalformed)
	}
	output.SemanticBeforeDigest, output.SemanticAfterDigest = before.digest, after.digest
	deltaText, added, removed := semanticDelta(before.facts, after.facts)
	output.ObservationCounts.AddedSemanticFacts = uint64(len(added))
	output.ObservationCounts.RemovedSemanticFacts = uint64(len(removed))
	if input.RegistryDigest == "" || !validDigest(input.RegistryDigest) {
		return finish(output, DecisionUnknown, ReasonRequiredInputMissing)
	}
	registry, issue := normalizeRegistry(input.Registry)
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	if registry.digest != input.RegistryDigest {
		return finish(output, DecisionFailClosed, ReasonDigestMismatch)
	}
	if issue := validateManifest(input, before.digest, after.digest, registry.digest); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	changed, issue := resolveChangedSurfaces(input.Changes, registry)
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	output.ChangedSurfaces = changed
	output.ObservationCounts.ChangedRegistered = uint64(len(changed))
	if issue := validateSourceBindings(input, before.digest, after.digest); issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	receipts, issue := validateReceipts(input, registry, changed, before, after, deltaText)
	output.ReceiptSurfaces = receipts.valid
	output.ObservationCounts.ValidReceipts = uint64(len(receipts.valid))
	if issue.decision != "" {
		return finish(output, issue.decision, issue.reason)
	}
	path := validatePath(input, registry, receipts.bySurface, before.digest, after.digest, deltaText)
	output.PathClosureDigest = path.digest
	output.ObservationCounts.PathEdges = path.counts.PathEdges
	output.ObservationCounts.PathClaims = path.counts.PathClaims
	output.ObservationCounts.PathEvidence = path.counts.PathEvidence
	output.ObservationCounts.CandidateObservations = path.counts.CandidateObservations
	output.ObservationCounts.AcceptedLifts = path.counts.AcceptedLifts
	if path.decision != "" {
		return finish(output, path.decision, path.reason)
	}
	if len(changed) > 0 {
		output.SemanticDeltaDigest = ""
		if deltaText != "" {
			output.SemanticDeltaDigest = digestBytes([]byte(deltaText))
		}
	}
	output.AcceptedSurfaces = append([]string(nil), changed...)
	return finish(output, DecisionPass, ReasonNone)
}

func finish(output Output, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	output.AcceptedSurfaces = sortedUnique(output.AcceptedSurfaces)
	output.ChangedSurfaces = sortedUnique(output.ChangedSurfaces)
	output.ReceiptSurfaces = sortedUnique(output.ReceiptSurfaces)
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}

func validateRequiredInput(input Input) oracleValidation {
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" ||
		input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" ||
		input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" ||
		len(input.Registry) == 0 || len(input.ResourceReceipts) == 0 || !input.Manifest.Complete ||
		(len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	if !validDigest(input.Config.ToolchainDigest) || !validDigest(input.Config.Profile.Digest) {
		return oracleValidation{DecisionUnknown, ReasonRequiredInputMissing}
	}
	return oracleValidation{}
}
