package artifactemit

import "encoding/json"

func decodeSymbolicReaderRequestSource(payload []byte) (SymbolicValueReaderProjection, bool) {
	var value SymbolicValueReaderProjection
	if err := json.Unmarshal(payload, &value); err != nil {
		return value, false
	}
	return value, true
}

func symbolicReaderRequestSourceValid(value SymbolicValueReaderProjection, subjectSHA string) bool {
	return value.Schema == symbolicReaderProjectionSchema &&
		value.SubjectSHA == subjectSHA && value.MetricID == symbolicReaderProjectionMetric &&
		value.Decision == "PASS" && value.Resolution == "READER_PROJECTION_ONLY" &&
		value.Reason == "CANONICAL_READER_PROJECTIONS_BOUND" &&
		value.Coordinates.Satisfied == 18 && value.Coordinates.Total == 18 &&
		value.Coordinates.BasisPoints == 10000 && len(value.Readers) == 3 &&
		symbolicReaderValidDigest(value.Digest) &&
		value.Digest == symbolicReaderProjectionDigest(value) &&
		symbolicReaderRequestUpstreamValid(value.Source)
}

func symbolicReaderRequestUpstreamValid(source SymbolicValueReaderProjectionSource) bool {
	return source.Schema == "gooo/symbolic-invocation-value-reachability/v1" &&
		source.MetricID == "gooo.metric.compiler.symbolic-value-reachability.v1" &&
		source.Decision == "PASS" && source.Resolution == "SCHEMA_VALUE_REACHABILITY_ONLY" &&
		symbolicReaderValidDigest(source.ReachabilityDigest) &&
		symbolicReaderValidDigest(source.FileDigest)
}

func symbolicReaderRequestReader(
	readers []SymbolicValueReaderProjectionView,
	audience string,
) (SymbolicValueReaderProjectionView, bool) {
	var selected SymbolicValueReaderProjectionView
	matches := 0
	for _, reader := range readers {
		if reader.Audience == audience {
			selected = reader
			matches++
		}
	}
	return selected, matches == 1
}

func symbolicReaderRequestIDsUnique(values []string) bool {
	if len(values) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}
