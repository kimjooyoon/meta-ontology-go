package pressureshadow

import (
	"encoding/json"
)

type Decision string

const (
	DecisionPass       Decision = "PASS"
	DecisionUnknown    Decision = "UNKNOWN"
	DecisionFailClosed Decision = "FAIL_CLOSED"
)

type Reason string

const (
	ReasonNone                 Reason = "NONE"
	ReasonInvalidInput         Reason = "INVALID_INPUT"
	ReasonRequiredInputMissing Reason = "REQUIRED_INPUT_MISSING"
	ReasonMissingPathCoverage  Reason = "MISSING_PATH_COVERAGE"
	ReasonOrphanPathCoverage   Reason = "ORPHAN_PATH_COVERAGE"
	ReasonBindingMismatch      Reason = "BINDING_MISMATCH"
)

type EnforcementEffect string

const EnforcementNoEffect EnforcementEffect = "NO_EFFECT"

type Result struct {
	Schema                 string            `json:"schema"`
	InputDigest            string            `json:"input_digest"`
	Decision               Decision          `json:"decision"`
	Reason                 Reason            `json:"reason"`
	MissingPathIDs         []string          `json:"missing_path_ids"`
	OrphanPathIDs          []string          `json:"orphan_path_ids"`
	MissingBindingPathIDs  []string          `json:"missing_binding_path_ids"`
	BindingMismatchPathIDs []string          `json:"binding_mismatch_path_ids"`
	EnforcementEffect      EnforcementEffect `json:"enforcement_effect"`
	ResultDigest           string            `json:"result_digest"`
	ReplayDigest           string            `json:"replay_digest"`
}

// Validate checks only the S1a2a path-row and outer tuple contract.
func Validate(input Input) Result {
	canonical, err := CanonicalInputBytes(input)
	if err != nil {
		return makeResult(CanonicalInputDigest(input), DecisionFailClosed, ReasonInvalidInput,
			nil, nil, nil, nil)
	}
	inputDigest := digestBytes(canonical)
	paths := selectorPathIDs(input)
	rows := coverageRows(input)
	missing := missingPathIDs(paths, rows)
	orphan := orphanPathIDs(paths, rows)
	missingBinding, mismatch := bindingIssues(input, paths, rows)
	decision, reason := pathDecision(len(paths), missing, orphan, missingBinding, mismatch)
	return makeResult(inputDigest, decision, reason, missing, orphan, missingBinding, mismatch)
}

func selectorPathIDs(input Input) map[string]struct{} {
	ids := make(map[string]struct{}, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		ids[pathID(path)] = struct{}{}
	}
	return ids
}

func coverageRows(input Input) map[string]PathCoverage {
	rows := make(map[string]PathCoverage, len(input.PathCoverage))
	for _, row := range input.PathCoverage {
		rows[row.PathID] = row
	}
	return rows
}

func missingPathIDs(paths map[string]struct{}, rows map[string]PathCoverage) []string {
	missing := []string{}
	for id := range paths {
		if _, exists := rows[id]; !exists {
			missing = append(missing, id)
		}
	}
	return missing
}

func orphanPathIDs(paths map[string]struct{}, rows map[string]PathCoverage) []string {
	orphan := []string{}
	for id := range rows {
		if _, exists := paths[id]; !exists {
			orphan = append(orphan, id)
		}
	}
	return orphan
}

func bindingIssues(input Input, paths map[string]struct{}, rows map[string]PathCoverage) ([]string, []string) {
	missing, mismatch := []string{}, []string{}
	selector := []string{input.Selector.SnapshotDigest, input.Selector.PolicyDigest, input.Selector.RegistryDigest}
	for id := range paths {
		row, exists := rows[id]
		if !exists {
			continue
		}
		values := []string{row.SnapshotDigest, row.PolicyDigest, row.RegistryDigest}
		blank, unequal := false, false
		for index := range selector {
			blank = blank || selector[index] == "" || values[index] == ""
			unequal = unequal || values[index] != "" && values[index] != selector[index]
		}
		if blank {
			missing = append(missing, id)
		}
		if unequal {
			mismatch = append(mismatch, id)
		}
	}
	return missing, mismatch
}

func pathDecision(pathCount int, missing, orphan, missingBinding, mismatch []string) (Decision, Reason) {
	switch {
	case len(orphan) > 0:
		return DecisionFailClosed, ReasonOrphanPathCoverage
	case pathCount == 0:
		return DecisionUnknown, ReasonRequiredInputMissing
	case len(missing) > 0:
		return DecisionUnknown, ReasonMissingPathCoverage
	case len(missingBinding) > 0:
		return DecisionUnknown, ReasonRequiredInputMissing
	case len(mismatch) > 0:
		return DecisionUnknown, ReasonBindingMismatch
	default:
		return DecisionPass, ReasonNone
	}
}

func makeResult(inputDigest string, decision Decision, reason Reason, missing, orphan,
	missingBinding, mismatch []string) Result {
	result := Result{
		Schema:                 SchemaVersion,
		InputDigest:            inputDigest,
		Decision:               decision,
		Reason:                 reason,
		MissingPathIDs:         sortedStrings(missing),
		OrphanPathIDs:          sortedStrings(orphan),
		MissingBindingPathIDs:  sortedStrings(missingBinding),
		BindingMismatchPathIDs: sortedStrings(mismatch),
		EnforcementEffect:      EnforcementNoEffect,
	}
	result.ResultDigest = CanonicalResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("replay\x00" + inputDigest + "\x00" + result.ResultDigest))
	return result
}

// CanonicalResultDigest binds the decision, reason, sets, and input digest.
func CanonicalResultDigest(result Result) string {
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
