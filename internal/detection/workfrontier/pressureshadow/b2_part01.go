package pressureshadow

const (
	ReasonUnregisteredPressureRecord    Reason = "UNREGISTERED_PRESSURE_RECORD"
	ReasonSelectorPressureMissing       Reason = "SELECTOR_PRESSURE_MISSING"
	ReasonRequiredPressureRecordMissing Reason = "REQUIRED_PRESSURE_RECORD_MISSING"
)

type B2Result struct {
	Schema                           string                     `json:"schema"`
	InputDigest                      string                     `json:"input_digest"`
	UpstreamResultDigest             string                     `json:"upstream_result_digest"`
	Decision                         Decision                   `json:"decision"`
	Reason                           Reason                     `json:"reason"`
	MissingRequiredPressureRecordIDs []RequiredPressureSetIssue `json:"missing_required_pressure_record_ids"`
	MissingSelectorPressureIDs       []RequiredPressureSetIssue `json:"missing_selector_pressure_ids"`
	UnregisteredPressureRecordIDs    []RequiredPressureSetIssue `json:"unregistered_pressure_record_ids"`
	EnforcementEffect                EnforcementEffect          `json:"enforcement_effect"`
	ResultDigest                     string                     `json:"result_digest"`
	ReplayDigest                     string                     `json:"replay_digest"`
}

// ValidateB2 checks per-path pressure-record registration and closure.
func ValidateB2(input Input) B2Result {
	return evaluateB2(input, ValidateB1(input))
}

// ValidateB2Bytes preserves B1's strict wire boundary before mapping.
func ValidateB2Bytes(data []byte) B2Result {
	upstream := ValidateB1Bytes(data)
	if upstream.Decision != DecisionPass {
		return fromB2Upstream(upstream)
	}
	input, err := DecodeInput(data)
	if err != nil {
		return finishB2(newB2Result(upstream), DecisionFailClosed, ReasonUpstreamFailClosed)
	}
	return evaluateB2(input, upstream)
}
func evaluateB2(input Input, upstream B1Result) B2Result {
	if upstream.Decision != DecisionPass {
		return fromB2Upstream(upstream)
	}
	missingRecords, missingSelector, unregistered := b2Issues(input)
	result := newB2Result(upstream)
	result.MissingRequiredPressureRecordIDs = missingRecords
	result.MissingSelectorPressureIDs = missingSelector
	result.UnregisteredPressureRecordIDs = unregistered
	switch {
	case len(unregistered) > 0:
		return finishB2(result, DecisionFailClosed, ReasonUnregisteredPressureRecord)
	case len(missingSelector) > 0:
		return finishB2(result, DecisionUnknown, ReasonSelectorPressureMissing)
	case len(missingRecords) > 0:
		return finishB2(result, DecisionUnknown, ReasonRequiredPressureRecordMissing)
	default:
		return finishB2(result, DecisionPass, ReasonNone)
	}
}
