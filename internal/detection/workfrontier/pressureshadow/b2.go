package pressureshadow

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

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

func b2Issues(input Input) ([]RequiredPressureSetIssue, []RequiredPressureSetIssue,
	[]RequiredPressureSetIssue) {
	selectorIDs := make([]string, 0, len(input.Selector.Pressures))
	for _, pressure := range input.Selector.Pressures {
		selectorIDs = append(selectorIDs, pressureID(pressure))
	}
	selectorIDs = sortedStrings(selectorIDs)
	paths := make(map[string][]string, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		paths[pathID(path)] = path.RequiredPressureIDs
	}
	rows := coverageRows(input)
	ids := make([]string, 0, len(paths))
	for id := range paths {
		ids = append(ids, id)
	}
	var missingRecords, missingSelector, unregistered []RequiredPressureSetIssue
	for _, id := range sortedStrings(ids) {
		recordIDs := pressureRecordIDs(rows[id].Coverage.PressureRecords)
		if values := pressureDifference(paths[id], recordIDs); len(values) > 0 {
			missingRecords = append(missingRecords, RequiredPressureSetIssue{id, values})
		}
		if values := pressureDifference(paths[id], selectorIDs); len(values) > 0 {
			missingSelector = append(missingSelector, RequiredPressureSetIssue{id, values})
		}
		if values := pressureDifference(recordIDs, selectorIDs); len(values) > 0 {
			unregistered = append(unregistered, RequiredPressureSetIssue{id, values})
		}
	}
	return missingRecords, missingSelector, unregistered
}

func pressureRecordIDs(records []pressurecoverage.PressureRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.PressureID)
	}
	return sortedStrings(ids)
}

func fromB2Upstream(upstream B1Result) B2Result {
	decision, reason := DecisionUnknown, ReasonUpstreamUnknown
	if upstream.Decision == DecisionFailClosed {
		decision, reason = DecisionFailClosed, ReasonUpstreamFailClosed
	}
	return finishB2(newB2Result(upstream), decision, reason)
}

func newB2Result(upstream B1Result) B2Result {
	return B2Result{
		Schema:               SchemaVersion,
		InputDigest:          upstream.InputDigest,
		UpstreamResultDigest: upstream.ResultDigest,
		EnforcementEffect:    EnforcementNoEffect,
	}
}

func finishB2(result B2Result, decision Decision, reason Reason) B2Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeB2Result(result)
	result.ResultDigest = CanonicalB2ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("b2-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}

// CanonicalB2ResultDigest binds upstream state and all three issue sets.
func CanonicalB2ResultDigest(result B2Result) string {
	result = normalizeB2Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func normalizeB2Result(result B2Result) B2Result {
	result.MissingRequiredPressureRecordIDs = normalizeSetIssues(result.MissingRequiredPressureRecordIDs)
	result.MissingSelectorPressureIDs = normalizeSetIssues(result.MissingSelectorPressureIDs)
	result.UnregisteredPressureRecordIDs = normalizeSetIssues(result.UnregisteredPressureRecordIDs)
	return result
}
