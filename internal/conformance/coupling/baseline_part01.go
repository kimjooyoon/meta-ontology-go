package coupling

// EvaluateBaseline is a deliberately separate full-suite implementation. It
// re-derives registry closure, semantic equality, receipt predicates, and
// resource bindings from the typed input without calling Evaluate or any of
// the oracle's private helpers.
func EvaluateBaseline(input Input) BaselineResult {
	result := BaselineResult{FullSuite: true, ObservationCounts: baselineCounts(input)}
	if !baselineResourceBindingsEqual(input.Config.ResourceBinding, input.ResourceRegistry) {
		return baselineUnknown(result, input, ReasonResourceUnbound)
	}
	resources, ok, reason := baselineResources(input.ResourceReceipts, input.Config.ResourceBinding)
	result.Resources = resources
	result.WorkUnits = resources.WorkUnits
	if !ok {
		return baselineUnknown(result, input, reason)
	}
	if input.Schema != SchemaV1 || input.Config.ToolchainDigest == "" || input.Config.Profile.ID == "" || input.Config.Profile.Version == "" || input.Config.Profile.Digest == "" || input.AuthoritySourceBefore == "" || input.AuthoritySourceAfter == "" || len(input.Registry) == 0 || !input.Manifest.Complete || (len(input.Changes) == 0 && !input.Manifest.ZeroChange) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	if !baselineDigest(input.Config.ToolchainDigest) || !baselineDigest(input.Config.Profile.Digest) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	before, ok := baselineSemantic(input.SemanticBefore)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	after, ok := baselineSemantic(input.SemanticAfter)
	if !ok {
		return baselineFail(result, input, ReasonPathMalformed)
	}
	registry, ok := baselineRegistry(input.Registry)
	if !ok || registry.digest != input.RegistryDigest {
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	if !baselineManifest(input, before.digest, after.digest, registry.digest) {
		if !input.Manifest.Complete {
			return baselineUnknown(result, input, ReasonRequiredInputMissing)
		}
		return baselineFail(result, input, ReasonDigestMismatch)
	}
	changed, ok := baselineChanged(input.Changes, registry.bySymbol)
	if !ok {
		return baselineFail(result, input, ReasonChangedSurface)
	}
	result.LocalizedSurfaces = append([]string(nil), changed...)
	if receiptsOK, receiptReason := baselineReceipts(input, registry.bySurface, changed, before.digest, after.digest); !receiptsOK {
		if receiptReason == ReasonMissingReceipt || receiptReason == ReasonStaleReceipt {
			return baselineUnknown(result, input, receiptReason)
		}
		return baselineFail(result, input, receiptReason)
	}
	delta := baselineDelta(before.facts, after.facts)
	if !baselineClaims(input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonInvalidDelta)
	}
	if len(changed) > 0 && !baselinePath(input, registry.bySurface, input.Receipts, before.digest, after.digest, delta) {
		return baselineFail(result, input, ReasonPathClosure)
	}
	result.Decision, result.Reason, result.LocalizedSurfaces = DecisionPass, ReasonNone, nil
	return result
}
func baselineResourceBindingsEqual(left, right ResourceBindingConfig) bool {
	return left.ProviderID != "" && left.ProviderID == right.ProviderID && left.ObserverID != "" && left.ObserverID == right.ObserverID && left.ProviderDigest != "" && left.ProviderDigest == right.ProviderDigest && left.ObserverDigest != "" && left.ObserverDigest == right.ObserverDigest && left.SnapshotDigest != "" && left.SnapshotDigest == right.SnapshotDigest && left.SourceDigest != "" && left.SourceDigest == right.SourceDigest
}
