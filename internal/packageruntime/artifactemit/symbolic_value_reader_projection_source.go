package artifactemit

func symbolicReaderSourceChecks(
	source SymbolicValueReachability,
	subjectSHA string,
) symbolicReaderChecks {
	return symbolicReaderChecks{
		Schema:     source.Schema == "gooo/symbolic-invocation-value-reachability/v1",
		Subject:    subjectSHA != "" && source.SubjectSHA == subjectSHA,
		Metric:     source.MetricID == "gooo.metric.compiler.symbolic-value-reachability.v1",
		Decision:   source.Decision == "PASS",
		Resolution: source.Resolution == "SCHEMA_VALUE_REACHABILITY_ONLY",
		InternalDigest: symbolicReaderValidDigest(source.Digest) &&
			source.Digest == symbolicReaderReachabilityDigest(source),
		UpstreamDigests: symbolicReaderValidDigest(source.Source.ArtifactDigest) &&
			symbolicReaderValidDigest(source.Source.ContractDigest),
		UnknownBranches:   source.Summary.UnknownPolicyBranches == 0,
		UniqueIndicatorID: symbolicReaderUniqueIndicatorIDs(source.Indicators),
	}
}

func symbolicReaderProjectionSource(
	source SymbolicValueReachability,
	payload []byte,
) SymbolicValueReaderProjectionSource {
	return SymbolicValueReaderProjectionSource{
		Schema:             source.Schema,
		MetricID:           source.MetricID,
		Decision:           source.Decision,
		Resolution:         source.Resolution,
		ReachabilityDigest: source.Digest,
		FileDigest:         symbolicReaderBytesDigest(payload),
	}
}

func symbolicReaderUniqueIndicatorIDs(indicators []SymbolicValueContractIndicator) bool {
	if len(indicators) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(indicators))
	for _, indicator := range indicators {
		if indicator.ID == "" {
			return false
		}
		if _, found := seen[indicator.ID]; found {
			return false
		}
		seen[indicator.ID] = struct{}{}
	}
	return true
}
