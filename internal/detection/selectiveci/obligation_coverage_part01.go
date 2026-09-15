package selectiveci

func EvaluateObligationCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return ObserveObligationCoverage(input)
}
func ObserveObligationCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	input = normalizeCoverageInput(input)
	result := coverageResultFor(input)
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return sealCoverage(result, CoverageDecisionUnknown, coverageInputReason(input))
	}
	result.InputDigest = digestBytes(canonical)
	if input.SchemaVersion != ObligationCoverageSchemaVersion {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonUnsupportedSchema)
	}
	if input.ChangedRootIDs == nil {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonMissingInput)
	}
	if !validDigest(input.SnapshotDigest) {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonInvalidSnapshot)
	}
	if reason := validateCoverageRegistry(input.Registry); reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	graph, err := input.Graph.Normalized()
	if err != nil {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonInvalidGraph)
	}
	result.GraphDigest = graph.Digest()
	if reason := validateCoverageBindings(graph, input.Registry); reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	if graph.SnapshotDigest != input.SnapshotDigest {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonStaleSnapshot)
	}
	if graph.RegistryDigest != input.Registry.Digest || graph.PolicyDigest != input.Registry.PolicyDigest {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonStaleGraph)
	}
	roots, reason := coverageRoots(input.ChangedRootIDs, graph)
	if reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	result.ChangedRootCount = uint64(len(roots))
	if len(roots) == 0 {
		return sealCoverage(result, CoverageDecisionExact, CoverageReasonNoChange)
	}
	required, uncovered := reachableCoverage(graph, input.Registry, roots)
	result.RequiredObligationCount = uint64(len(required))
	result.UncoveredRootIDs = uncovered
	result.UncoveredChangedRootCount = uint64(len(uncovered))
	result.CoveredChangedRootCount = result.ChangedRootCount - result.UncoveredChangedRootCount
	bound, scanned, reason := coverageCommandRecords(required, input.Registry)
	result.BoundCommandCount = bound
	work, ok := coverageWorkUnits(result.ChangedRootCount, result.RequiredObligationCount, scanned)
	if len(uncovered) != 0 {
		if ok {
			result.DeterministicWorkUnits = work
		}
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonMissingObligation)
	}
	if !ok {
		return sealCoverage(result, CoverageDecisionUnknown, CoverageReasonWorkOverflow)
	}
	result.DeterministicWorkUnits = work
	if reason != "" {
		return sealCoverage(result, CoverageDecisionUnknown, reason)
	}
	result.RequiredObligationIDs = required
	return sealCoverage(result, CoverageDecisionExact, CoverageReasonComplete)
}
func EvaluateCoverage(input ObligationCoverageInput) ObligationCoverageResult {
	return EvaluateObligationCoverage(input)
}
