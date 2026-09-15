package impactcoverage

func resultFor(input Input) Result {
	result := Result{Schema: SchemaV1, ChangedStableIDs: []string{}, UncoveredPaths: []string{}}
	if input.Base == nil || input.Head == nil {
		return result
	}
	result.BaseSnapshotDigest = input.Base.Digest
	result.HeadSnapshotDigest = input.Head.Digest
	result.BaseSourceMapDigest = input.Base.SourceMapDigest
	result.HeadSourceMapDigest = input.Head.SourceMapDigest
	result.BaseRegistryDigest = input.Base.RegistryDigest
	result.HeadRegistryDigest = input.Head.RegistryDigest
	return result
}
func seal(result Result, decision Decision, reason Reason) Result {
	result.Decision = decision
	result.Reason = reason
	result.FullSuiteRequired = decision == DecisionUnknown
	if decision == DecisionUnknown {
		result.ChangedStableIDs = []string{}
	}
	result.ChangedStableIDs = sortedUnique(result.ChangedStableIDs)
	result.UncoveredPaths = sortedUnique(result.UncoveredPaths)
	result.OutputDigest = ""
	canonical, err := result.CanonicalJSON()
	if err == nil {
		result.OutputDigest = digestBytes(canonical)
	}
	return result
}
