package pressureshadow

import (
	"encoding/json"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

const (
	ReasonPressureCoverageFailClosed Reason = "PRESSURE_COVERAGE_FAIL_CLOSED"
	ReasonPressureCoverageUnknown    Reason = "PRESSURE_COVERAGE_UNKNOWN"
)

type S1B1PathObservation struct {
	PathID string                  `json:"path_id"`
	Result pressurecoverage.Result `json:"result"`
}

type S1B1Result struct {
	Schema                         string                `json:"schema"`
	InputDigest                    string                `json:"input_digest"`
	UpstreamResultDigest           string                `json:"upstream_result_digest"`
	Decision                       Decision              `json:"decision"`
	Reason                         Reason                `json:"reason"`
	A2Observations                 []S1B1PathObservation `json:"a2_observations"`
	PressureCoveragePassPathIDs    []string              `json:"pressure_coverage_pass_path_ids"`
	PressureCoverageUnknownPathIDs []string              `json:"pressure_coverage_unknown_path_ids"`
	PressureCoverageFailPathIDs    []string              `json:"pressure_coverage_fail_path_ids"`
	EnforcementEffect              EnforcementEffect     `json:"enforcement_effect"`
	ResultDigest                   string                `json:"result_digest"`
	ReplayDigest                   string                `json:"replay_digest"`
}

func ValidateS1B1(input Input) S1B1Result {
	return evaluateS1B1(input, ValidateB2(input))
}

func ValidateS1B1Bytes(data []byte) S1B1Result {
	upstream := ValidateB2Bytes(data)
	if upstream.Decision != DecisionPass {
		return evaluateS1B1(Input{}, upstream)
	}
	input, err := DecodeInput(data)
	if err != nil {
		return finishS1B1(newS1B1Result(upstream), DecisionFailClosed, ReasonUpstreamFailClosed)
	}
	return evaluateS1B1(input, upstream)
}

func evaluateS1B1(input Input, upstream B2Result) S1B1Result {
	if upstream.Decision != DecisionPass {
		decision, reason := DecisionUnknown, ReasonUpstreamUnknown
		if upstream.Decision == DecisionFailClosed {
			decision, reason = DecisionFailClosed, ReasonUpstreamFailClosed
		}
		return finishS1B1(newS1B1Result(upstream), decision, reason)
	}
	result, rows := newS1B1Result(upstream), coverageRows(input)
	ids := make([]string, len(input.Selector.Paths))
	for index, path := range input.Selector.Paths {
		ids[index] = pathID(path)
	}
	for _, id := range sortedStrings(ids) {
		observed := pressurecoverage.Evaluate(rows[id].Coverage)
		result.A2Observations = append(result.A2Observations, S1B1PathObservation{id, observed})
		switch observed.Decision {
		case pressurecoverage.DecisionFailClosed:
			result.PressureCoverageFailPathIDs = append(result.PressureCoverageFailPathIDs, id)
		case pressurecoverage.DecisionUnknown:
			result.PressureCoverageUnknownPathIDs = append(result.PressureCoverageUnknownPathIDs, id)
		default:
			result.PressureCoveragePassPathIDs = append(result.PressureCoveragePassPathIDs, id)
		}
	}
	switch {
	case len(result.PressureCoverageFailPathIDs) > 0:
		return finishS1B1(result, DecisionFailClosed, ReasonPressureCoverageFailClosed)
	case len(result.PressureCoverageUnknownPathIDs) > 0:
		return finishS1B1(result, DecisionUnknown, ReasonPressureCoverageUnknown)
	default:
		return finishS1B1(result, DecisionPass, ReasonNone)
	}
}

func newS1B1Result(upstream B2Result) S1B1Result {
	return S1B1Result{Schema: SchemaVersion, InputDigest: upstream.InputDigest,
		UpstreamResultDigest: upstream.ResultDigest, EnforcementEffect: EnforcementNoEffect}
}

func finishS1B1(result S1B1Result, decision Decision, reason Reason) S1B1Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeS1B1Result(result)
	result.ResultDigest = CanonicalS1B1ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("s1b1-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}

func CanonicalS1B1ResultDigest(result S1B1Result) string {
	result = normalizeS1B1Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func normalizeS1B1Result(result S1B1Result) S1B1Result {
	result.A2Observations = append([]S1B1PathObservation{}, result.A2Observations...)
	sort.Slice(result.A2Observations, func(left, right int) bool {
		return result.A2Observations[left].PathID < result.A2Observations[right].PathID
	})
	result.PressureCoveragePassPathIDs = sortedStrings(result.PressureCoveragePassPathIDs)
	result.PressureCoverageUnknownPathIDs = sortedStrings(result.PressureCoverageUnknownPathIDs)
	result.PressureCoverageFailPathIDs = sortedStrings(result.PressureCoverageFailPathIDs)
	return result
}
