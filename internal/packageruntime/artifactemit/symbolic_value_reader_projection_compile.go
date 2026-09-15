package artifactemit

import (
	"encoding/json"
	"strings"
)

func CompileSymbolicValueReaderProjection(
	reachabilityJSON []byte,
	subjectSHA string,
) (SymbolicValueReaderProjection, error) {
	var reachability SymbolicValueReachability
	if err := json.Unmarshal(reachabilityJSON, &reachability); err != nil {
		return SymbolicValueReaderProjection{}, err
	}
	checks := symbolicReaderSourceChecks(reachability, subjectSHA)
	readers := symbolicReaderBuildViews(reachability, &checks)
	indicators := symbolicReaderIndicators(checks)
	passed := checks.passed()
	decision, resolution, reason := symbolicReaderDecision(indicators, passed)
	if !passed {
		symbolicReaderLowerResolution(readers)
	}
	result := SymbolicValueReaderProjection{
		Schema: symbolicReaderProjectionSchema, SubjectSHA: subjectSHA,
		MetricID: symbolicReaderProjectionMetric, Decision: decision,
		Resolution: resolution, Reason: reason,
		Source:             symbolicReaderProjectionSource(reachability, reachabilityJSON),
		Readers:            readers,
		Coordinates:        symbolicReaderMetricCoordinates(indicators),
		Classes:            symbolicReaderClasses(indicators),
		Indicators:         indicators,
		Views:              symbolicReaderMetricViews(indicators),
		Proofs:             symbolicReaderProofs(indicators),
		Effects:            SymbolicValueContractEffects{RepositoryWrites: 0, MutationAuthority: false},
		PromotionCreditBPS: 0,
		NotClaimed: []string{
			"reader comprehension", "runtime frequency", "external user adoption",
			"domain correctness", "production readiness", "access-control enforcement",
		},
	}
	result.Digest = symbolicReaderProjectionDigest(result)
	return result, nil
}

func symbolicReaderDecision(
	indicators []SymbolicValueContractIndicator,
	passed bool,
) (string, string, string) {
	if passed {
		return "PASS", "READER_PROJECTION_ONLY", "CANONICAL_READER_PROJECTIONS_BOUND"
	}
	for _, indicator := range indicators {
		if !indicator.Satisfied {
			name := strings.NewReplacer(".", "_", "-", "_").Replace(indicator.ID)
			return "FAIL_CLOSED", "INVARIANT_ONLY", strings.ToUpper(name) + "_FAILED"
		}
	}
	return "FAIL_CLOSED", "INVARIANT_ONLY", "READER_PROJECTION_UNKNOWN_FAILURE"
}
