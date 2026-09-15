package pressurecoverage

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionFailClosed Decision = "FAIL_CLOSED"
)

type Reason string

const (
	ReasonNone                         Reason = "NONE"
	ReasonInvalidInput                 Reason = "INVALID_INPUT"
	ReasonRequiredInputMissing         Reason = "REQUIRED_INPUT_MISSING"
	ReasonSnapshotMismatch             Reason = "SNAPSHOT_MISMATCH"
	ReasonPolicyFloorViolation         Reason = "POLICY_FLOOR_VIOLATION"
	ReasonPressureCardinalityShortfall Reason = "PRESSURE_CARDINALITY_SHORTFALL"
	ReasonApplicabilityOrGroupUnproven Reason = "APPLICABILITY_OR_GROUP_UNPROVEN"
	ReasonIndependentGroupShortfall    Reason = "INDEPENDENT_GROUP_SHORTFALL"
)

type Result struct {
	Schema                string   `json:"schema"`
	InputDigest           string   `json:"input_digest"`
	RequiredPressureCount uint64   `json:"required_pressure_count"`
	DistinctGroupCount    uint64   `json:"distinct_group_count"`
	RequiredPressureIDs   []string `json:"required_pressure_ids"`
	RequiredGroupIDs      []string `json:"required_group_ids"`
	MissingPressureIDs    []string `json:"missing_pressure_ids"`
	Decision              Decision `json:"decision"`
	Reason                Reason   `json:"reason"`
	ResultDigest          string   `json:"result_digest"`
	ReplayDigest          string   `json:"replay_digest"`
}

const a2PolicyFloor uint64 = 2

// Evaluate applies only A2 pressure-coverage semantics to an A1 Input.
func Evaluate(input Input) Result {
	result := newResult(input)
	if _, err := CanonicalInputBytes(input); err != nil {
		return finish(result, DecisionFailClosed, ReasonInvalidInput)
	}
	if blankBinding(input) {
		return finish(result, DecisionUnknown, ReasonRequiredInputMissing)
	}
	if !bindingMatches(input) {
		return finish(result, DecisionUnknown, ReasonSnapshotMismatch)
	}
	if input.RequestedK == 0 || input.MinimumIndependent == 0 {
		return finish(result, DecisionUnknown, ReasonRequiredInputMissing)
	}
	if input.RequestedK < a2PolicyFloor || input.MinimumIndependent < a2PolicyFloor ||
		input.MinimumIndependent > input.RequestedK {
		return finish(result, DecisionFailClosed, ReasonPolicyFloorViolation)
	}
	if len(input.RequiredPressureIDs) == 0 {
		return finish(result, DecisionUnknown, ReasonRequiredInputMissing)
	}
	return evaluateCoverage(result, input)
}
