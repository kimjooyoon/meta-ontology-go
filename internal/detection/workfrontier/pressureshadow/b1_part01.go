package pressureshadow

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
