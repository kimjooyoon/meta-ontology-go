package pressureshadow

import (
	"encoding/json"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

const (
	ReasonSelectorSelectedFailedPressureCoverage  Reason = "SELECTOR_SELECTED_FAILED_PRESSURE_COVERAGE"
	ReasonSelectorSelectedUnknownPressureCoverage Reason = "SELECTOR_SELECTED_UNKNOWN_PRESSURE_COVERAGE"
	ReasonSelectorUnknown                         Reason = "SELECTOR_UNKNOWN"
)

type S1B2Result struct {
	Schema                       string               `json:"schema"`
	InputDigest                  string               `json:"input_digest"`
	UpstreamResultDigest         string               `json:"upstream_result_digest"`
	SelectorObserved             bool                 `json:"selector_observed"`
	SelectorResult               *workfrontier.Result `json:"selector_result"`
	SelectorResultDigest         string               `json:"selector_result_digest"`
	UnsafeSelectedUnknownPathIDs []string             `json:"unsafe_selected_unknown_path_ids"`
	UnsafeSelectedFailPathIDs    []string             `json:"unsafe_selected_fail_path_ids"`
	Decision                     Decision             `json:"decision"`
	Reason                       Reason               `json:"reason"`
	ExecutionAuthorized          bool                 `json:"execution_authorized"`
	EnforcementEffect            EnforcementEffect    `json:"enforcement_effect"`
	ResultDigest                 string               `json:"result_digest"`
	ReplayDigest                 string               `json:"replay_digest"`
}

// ValidateS1B2 correlates selector output with read-only S1b1 observations.
func ValidateS1B2(input Input) S1B2Result {
	return evaluateS1B2(input, ValidateS1B1(input))
}

// ValidateS1B2Bytes preserves S1b1's strict boundary before selector use.
func ValidateS1B2Bytes(data []byte) S1B2Result {
	upstream := ValidateS1B1Bytes(data)
	if structuralS1B1(upstream) {
		return finishS1B2(newS1B2Result(upstream), upstream.Decision, upstream.Reason)
	}
	input, err := DecodeInput(data)
	if err != nil {
		return finishS1B2(newS1B2Result(upstream), DecisionFailClosed, ReasonUpstreamFailClosed)
	}
	return evaluateS1B2(input, upstream)
}

func evaluateS1B2(input Input, upstream S1B1Result) S1B2Result {
	result := newS1B2Result(upstream)
	if structuralS1B1(upstream) {
		return finishS1B2(result, upstream.Decision, upstream.Reason)
	}
	selector := workfrontier.Select(input.Selector)
	result.SelectorObserved = true
	result.SelectorResult = &selector
	result.SelectorResultDigest = canonicalSelectorDigest(selector)
	result.UnsafeSelectedFailPathIDs = intersectSorted(selector.SelectedIDs,
		upstream.PressureCoverageFailPathIDs)
	result.UnsafeSelectedUnknownPathIDs = intersectSorted(selector.SelectedIDs,
		upstream.PressureCoverageUnknownPathIDs)
	decision, reason := s1b2Decision(result, upstream)
	return finishS1B2(result, decision, reason)
}

func structuralS1B1(result S1B1Result) bool {
	return result.Reason == ReasonUpstreamFailClosed || result.Reason == ReasonUpstreamUnknown
}

func s1b2Decision(result S1B2Result, upstream S1B1Result) (Decision, Reason) {
	if len(result.UnsafeSelectedFailPathIDs) > 0 {
		return DecisionFailClosed, ReasonSelectorSelectedFailedPressureCoverage
	}
	if len(upstream.PressureCoverageFailPathIDs) > 0 {
		return DecisionFailClosed, ReasonPressureCoverageFailClosed
	}
	if len(result.UnsafeSelectedUnknownPathIDs) > 0 {
		return DecisionUnknown, ReasonSelectorSelectedUnknownPressureCoverage
	}
	if len(upstream.PressureCoverageUnknownPathIDs) > 0 {
		return DecisionUnknown, ReasonPressureCoverageUnknown
	}
	if result.SelectorResult != nil && result.SelectorResult.Status == workfrontier.DecisionUnknown {
		return DecisionUnknown, ReasonSelectorUnknown
	}
	return DecisionPass, ReasonNone
}

func newS1B2Result(upstream S1B1Result) S1B2Result {
	return S1B2Result{
		Schema: SchemaVersion, InputDigest: upstream.InputDigest,
		UpstreamResultDigest: upstream.ResultDigest, UnsafeSelectedUnknownPathIDs: []string{},
		UnsafeSelectedFailPathIDs: []string{}, EnforcementEffect: EnforcementNoEffect,
	}
}

func finishS1B2(result S1B2Result, decision Decision, reason Reason) S1B2Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeS1B2Result(result)
	result.ResultDigest = CanonicalS1B2ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("s1b2-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}

func canonicalSelectorDigest(result workfrontier.Result) string {
	data, _ := json.Marshal(result)
	return digestBytes(append([]byte("s1b2-selector-result\x00"), data...))
}

// CanonicalS1B2ResultDigest binds selector evidence and both unsafe sets.
func CanonicalS1B2ResultDigest(result S1B2Result) string {
	result = normalizeS1B2Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}

func normalizeS1B2Result(result S1B2Result) S1B2Result {
	result.UnsafeSelectedUnknownPathIDs = sortedStrings(result.UnsafeSelectedUnknownPathIDs)
	result.UnsafeSelectedFailPathIDs = sortedStrings(result.UnsafeSelectedFailPathIDs)
	return result
}

func intersectSorted(selected, unsafe []string) []string {
	known := make(map[string]struct{}, len(unsafe))
	for _, id := range unsafe {
		known[id] = struct{}{}
	}
	result := make([]string, 0, len(selected))
	for _, id := range selected {
		if _, ok := known[id]; ok {
			result = append(result, id)
		}
	}
	result = sortedStrings(result)
	if len(result) == 0 {
		return []string{}
	}
	return result
}
