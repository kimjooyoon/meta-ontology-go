package transformationeffect

import (
	"sort"
	"strings"
)

func splitGoReportPolicy(root map[string]any, ids []string, candidates map[string][]splitGoIndicatorCandidate, unexpected map[string]struct{}) (bool, []string) {
	topDecision, _ := splitGoDirectString(root, "decision")
	topDecision = strings.ToUpper(topDecision)
	forceUnknown := false
	reasons := make([]string, 0)
	if topDecision != "PASS" && topDecision != "FAIL" && topDecision != "BLOCK" {
		forceUnknown = true
		reasons = append(reasons, "TOP_LEVEL_DECISION_UNKNOWN")
	}
	if len(unexpected) != 0 {
		forceUnknown = true
		reasons = append(reasons, "INDICATOR_SET_MISMATCH")
	}
	for _, id := range ids {
		if len(candidates[id]) != 1 {
			forceUnknown = true
			reasons = append(reasons, "INDICATOR_OBSERVATION_NOT_UNIQUE:"+id)
		}
	}
	if forceUnknown {
		sort.Strings(reasons)
		return true, reasons
	}

	passCount, failCount, unknownCount := splitGoVerdictCounts(ids, candidates)
	if topDecision == "PASS" && passCount != len(ids) {
		forceUnknown = true
		reasons = append(reasons, "PASS_REPORT_CONTAINS_NON_PASS")
	}
	if topDecision == "FAIL" && failCount == 0 {
		forceUnknown = true
		reasons = append(reasons, "FAIL_REPORT_WITHOUT_FAILED_INDICATOR")
	}
	if topDecision == "BLOCK" && unknownCount == 0 {
		forceUnknown = true
		reasons = append(reasons, "BLOCK_REPORT_WITHOUT_UNKNOWN_INDICATOR")
	}
	for _, id := range ids {
		if normalizeSplitGoVerdict(candidates[id][0].verdict) == "UNKNOWN" {
			reasons = append(reasons, "INDICATOR_DECISION_UNKNOWN:"+id)
		}
	}
	sort.Strings(reasons)
	return forceUnknown, reasons
}

func splitGoVerdictCounts(ids []string, candidates map[string][]splitGoIndicatorCandidate) (int, int, int) {
	var pass, fail, unknown int
	for _, id := range ids {
		switch normalizeSplitGoVerdict(candidates[id][0].verdict) {
		case "PASS":
			pass++
		case "FAIL":
			fail++
		default:
			unknown++
		}
	}
	return pass, fail, unknown
}
