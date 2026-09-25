package pressureshadow

import (
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
