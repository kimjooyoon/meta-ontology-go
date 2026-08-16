package pressureshadow

import (
	"encoding/json"
	"sort"
)

const (
	ReasonUpstreamFailClosed Reason = "UPSTREAM_FAIL_CLOSED"
	ReasonUpstreamUnknown    Reason = "UPSTREAM_UNKNOWN"
	ReasonRequiredSetMissing Reason = "REQUIRED_SET_MISSING"
	ReasonRequiredSetExtra   Reason = "REQUIRED_SET_EXTRA"
	ReasonRequestedKMissing  Reason = "REQUESTED_K_MISSING"
	ReasonRequestedKMismatch Reason = "REQUESTED_K_MISMATCH"
)

type RequiredPressureSetIssue struct {
	PathID      string   `json:"path_id"`
	PressureIDs []string `json:"pressure_ids"`
}

type RequestedKIssue struct {
	PathID    string `json:"path_id"`
	SelectorK uint64 `json:"selector_K"`
	CoverageK uint64 `json:"coverage_K"`
}

type B1Result struct {
	Schema                     string                     `json:"schema"`
	InputDigest                string                     `json:"input_digest"`
	UpstreamResultDigest       string                     `json:"upstream_result_digest"`
	Decision                   Decision                   `json:"decision"`
	Reason                     Reason                     `json:"reason"`
	MissingRequiredPressureIDs []RequiredPressureSetIssue `json:"missing_required_pressure_ids"`
	ExtraRequiredPressureIDs   []RequiredPressureSetIssue `json:"extra_required_pressure_ids"`
	MissingKPathIDs            []string                   `json:"missing_k_path_ids"`
	RequestedKIssues           []RequestedKIssue          `json:"requested_k_issues"`
	EnforcementEffect          EnforcementEffect          `json:"enforcement_effect"`
	ResultDigest               string                     `json:"result_digest"`
	ReplayDigest               string                     `json:"replay_digest"`
}

// ValidateB1 checks only required-pressure sets and per-path requested K.
func ValidateB1(input Input) B1Result {
	return evaluateB1(input, Validate(input))
}

// ValidateB1Bytes preserves the strict A2a wire boundary before mapping.
func ValidateB1Bytes(data []byte) B1Result {
	upstream := ValidateBytes(data)
	if upstream.Decision != DecisionPass {
		return fromUpstream(upstream)
	}
	input, err := DecodeInput(data)
	if err != nil {
		return finishB1(newB1Result(upstream), DecisionFailClosed, ReasonUpstreamFailClosed)
	}
	return evaluateB1(input, upstream)
}

func evaluateB1(input Input, upstream Result) B1Result {
	if upstream.Decision != DecisionPass {
		return fromUpstream(upstream)
	}
	missing, extra, missingK, mismatches := b1Issues(input)
	result := newB1Result(upstream)
	result.MissingRequiredPressureIDs = missing
	result.ExtraRequiredPressureIDs = extra
	result.MissingKPathIDs = missingK
	result.RequestedKIssues = mismatches
	switch {
	case len(extra) > 0:
		return finishB1(result, DecisionFailClosed, ReasonRequiredSetExtra)
	case len(mismatches) > 0:
		return finishB1(result, DecisionFailClosed, ReasonRequestedKMismatch)
	case len(missing) > 0:
		return finishB1(result, DecisionUnknown, ReasonRequiredSetMissing)
	case len(missingK) > 0:
		return finishB1(result, DecisionUnknown, ReasonRequestedKMissing)
	default:
		return finishB1(result, DecisionPass, ReasonNone)
	}
}

func b1Issues(input Input) ([]RequiredPressureSetIssue, []RequiredPressureSetIssue,
	[]string, []RequestedKIssue) {
	paths := make(map[string][]string, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		paths[pathID(path)] = path.RequiredPressureIDs
	}
	rows := coverageRows(input)
	missing, extra := []RequiredPressureSetIssue{}, []RequiredPressureSetIssue{}
	missingK, mismatches := []string{}, []RequestedKIssue{}
	ids := make([]string, 0, len(paths))
	for id := range paths {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		row := rows[id]
		if ids := pressureDifference(paths[id], row.Coverage.RequiredPressureIDs); len(ids) > 0 {
			missing = append(missing, RequiredPressureSetIssue{PathID: id, PressureIDs: ids})
		}
		if ids := pressureDifference(row.Coverage.RequiredPressureIDs, paths[id]); len(ids) > 0 {
			extra = append(extra, RequiredPressureSetIssue{PathID: id, PressureIDs: ids})
		}
		selectorK, coverageK := uint64(input.Selector.MinimumSelectedPressures), row.Coverage.RequestedK
		if selectorK == 0 || coverageK == 0 {
			missingK = append(missingK, id)
		} else if selectorK != coverageK {
			mismatches = append(mismatches, RequestedKIssue{id, selectorK, coverageK})
		}
	}
	return missing, extra, missingK, mismatches
}

func pressureDifference(left, right []string) []string {
	known := make(map[string]struct{}, len(right))
	for _, id := range right {
		known[id] = struct{}{}
	}
	result := []string{}
	for _, id := range left {
		if _, exists := known[id]; !exists {
			result = append(result, id)
		}
	}
	return sortedStrings(result)
}

func fromUpstream(upstream Result) B1Result {
	decision, reason := DecisionUnknown, ReasonUpstreamUnknown
	if upstream.Decision == DecisionFailClosed {
		decision, reason = DecisionFailClosed, ReasonUpstreamFailClosed
	}
	return finishB1(newB1Result(upstream), decision, reason)
}

func newB1Result(upstream Result) B1Result {
	return B1Result{
		Schema:                     SchemaVersion,
		InputDigest:                upstream.InputDigest,
		UpstreamResultDigest:       upstream.ResultDigest,
		MissingRequiredPressureIDs: []RequiredPressureSetIssue{},
		ExtraRequiredPressureIDs:   []RequiredPressureSetIssue{},
		MissingKPathIDs:            []string{},
		RequestedKIssues:           []RequestedKIssue{},
		EnforcementEffect:          EnforcementNoEffect,
	}
}

func finishB1(result B1Result, decision Decision, reason Reason) B1Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeB1Result(result)
	result.ResultDigest = CanonicalB1ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("b1-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}

// CanonicalB1ResultDigest binds the upstream result and every mapping issue.
func CanonicalB1ResultDigest(result B1Result) string {
	result = normalizeB1Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func normalizeB1Result(result B1Result) B1Result {
	result.MissingRequiredPressureIDs = normalizeSetIssues(result.MissingRequiredPressureIDs)
	result.ExtraRequiredPressureIDs = normalizeSetIssues(result.ExtraRequiredPressureIDs)
	result.MissingKPathIDs = sortedStrings(result.MissingKPathIDs)
	result.RequestedKIssues = normalizeKIssues(result.RequestedKIssues)
	return result
}

func normalizeSetIssues(issues []RequiredPressureSetIssue) []RequiredPressureSetIssue {
	result := append([]RequiredPressureSetIssue{}, issues...)
	for index := range result {
		result[index].PressureIDs = sortedStrings(result[index].PressureIDs)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].PathID < result[right].PathID
	})
	return result
}

func normalizeKIssues(issues []RequestedKIssue) []RequestedKIssue {
	result := append([]RequestedKIssue{}, issues...)
	sort.Slice(result, func(left, right int) bool {
		if result[left].PathID != result[right].PathID {
			return result[left].PathID < result[right].PathID
		}
		if result[left].SelectorK != result[right].SelectorK {
			return result[left].SelectorK < result[right].SelectorK
		}
		return result[left].CoverageK < result[right].CoverageK
	})
	return result
}
