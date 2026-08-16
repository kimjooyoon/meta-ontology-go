package pressurecoverage

import (
	"encoding/json"
	"sort"
)

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

func evaluateCoverage(result Result, input Input) Result {
	records := recordsByID(input.PressureRecords)
	groups := map[string]struct{}{}
	for _, id := range input.RequiredPressureIDs {
		record, ok := records[id]
		if !ok {
			result.MissingPressureIDs = append(result.MissingPressureIDs, id)
			continue
		}
		if record.IndependenceGroupID != "" {
			groups[record.IndependenceGroupID] = struct{}{}
		}
	}
	result.RequiredGroupIDs = groupIDs(groups)
	result.DistinctGroupCount = uint64(len(result.RequiredGroupIDs))
	if len(result.MissingPressureIDs) != 0 {
		return finish(result, DecisionUnknown, ReasonRequiredInputMissing)
	}
	if uint64(len(input.RequiredPressureIDs)) < input.RequestedK {
		return finish(result, DecisionUnknown, ReasonPressureCardinalityShortfall)
	}
	for _, id := range input.RequiredPressureIDs {
		record := records[id]
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return finish(result, DecisionUnknown, ReasonApplicabilityOrGroupUnproven)
		}
	}
	if result.DistinctGroupCount < input.MinimumIndependent {
		return finish(result, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	return finish(result, DecisionPass, ReasonNone)
}

func newResult(input Input) Result {
	return Result{
		Schema:                SchemaVersion,
		InputDigest:           CanonicalInputDigest(input),
		RequiredPressureCount: uint64(len(input.RequiredPressureIDs)),
		RequiredPressureIDs:   resultIDs(input.RequiredPressureIDs),
	}
}

func blankBinding(input Input) bool {
	return input.AuthoritySnapshotDigest == "" || input.PolicyDigest == "" ||
		input.RegistryDigest == "" || input.ToolchainOptionsDigest == ""
}

func bindingMatches(input Input) bool {
	return input.AuthoritySnapshotDigest == authorityBindingDigest(input, "authority-snapshot") &&
		input.PolicyDigest == authorityBindingDigest(input, "policy") &&
		input.RegistryDigest == authorityBindingDigest(input, "registry") &&
		input.ToolchainOptionsDigest == authorityBindingDigest(input, "toolchain-options")
}

func recordsByID(records []PressureRecord) map[string]PressureRecord {
	result := make(map[string]PressureRecord, len(records))
	for _, record := range records {
		result[record.PressureID] = record
	}
	return result
}

func groupIDs(groups map[string]struct{}) []string {
	result := make([]string, 0, len(groups))
	for group := range groups {
		result = append(result, group)
	}
	sort.Strings(result)
	return result
}

func resultIDs(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	if len(result) == 0 {
		return []string{}
	}
	return result
}

func CanonicalResultDigest(result Result) string {
	result.RequiredPressureIDs = resultIDs(result.RequiredPressureIDs)
	result.RequiredGroupIDs = resultIDs(result.RequiredGroupIDs)
	result.MissingPressureIDs = resultIDs(result.MissingPressureIDs)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func finish(result Result, decision Decision, reason Reason) Result {
	result.Decision, result.Reason = decision, reason
	result.RequiredPressureIDs = resultIDs(result.RequiredPressureIDs)
	result.RequiredGroupIDs = resultIDs(result.RequiredGroupIDs)
	result.MissingPressureIDs = resultIDs(result.MissingPressureIDs)
	result.ResultDigest = CanonicalResultDigest(result)
	result.ReplayDigest = digestBytes([]byte(result.InputDigest + "\x00" + result.ResultDigest))
	return result
}
