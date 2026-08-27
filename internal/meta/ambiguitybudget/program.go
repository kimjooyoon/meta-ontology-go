package ambiguitybudget

import (
	"fmt"
	"strings"
)

const programPrefix = "ambiguity-budget"

type computesProgram struct {
	Activity string
	Text     string
	Kind     string
	ID       string
}

func parseComputesProgram(activity, text string) (computesProgram, error) {
	if activity == "" || text == "" || text != strings.TrimSpace(text) || strings.ContainsAny(text, " \t\r\n") {
		return computesProgram{}, fmt.Errorf("non-canonical computes program")
	}
	parts := strings.Split(text, ":")
	if len(parts) == 3 && parts[0] == programPrefix && parts[1] == "budget-policy" && parts[2] != "" {
		return computesProgram{Activity: activity, Text: text, Kind: "BUDGET", ID: parts[2]}, nil
	}
	if len(parts) != 4 || parts[0] != programPrefix || parts[1] != "case" || parts[2] == "" || parts[3] == "" {
		return computesProgram{}, fmt.Errorf("unsupported computes program %q", text)
	}
	return computesProgram{Activity: activity, Text: text, Kind: "CASE", ID: parts[2]}, nil
}

func formatComputesProgram(program computesProgram) string {
	if program.Kind == "BUDGET" {
		return programPrefix + ":budget-policy:" + program.ID
	}
	return programPrefix + ":case:" + program.ID
}

func budgetBinding(policy BudgetPolicy) string {
	return programPrefix + ":budget-policy:" + policy.Version
}

func policyCounts(policy BudgetPolicy) IntegerSet {
	var counts IntegerSet
	for _, dimension := range policy.Dimensions {
		switch dimension.ID {
		case "interpretation_candidates":
			counts.InterpretationCandidates = dimension.Limit
		case "unresolved_branches":
			counts.UnresolvedBranches = dimension.Limit
		case "evidence_paths":
			counts.EvidencePaths = dimension.Limit
		}
	}
	return counts
}

func validPolicy(policy BudgetPolicy) bool {
	if policy.Schema != PolicySchema || policy.ID == "" || policy.Version == "" || policy.Authority != "CONTRACT_POLICY" || len(policy.Dimensions) != IntegerDimensions {
		return false
	}
	seen := map[string]bool{}
	for _, dimension := range policy.Dimensions {
		if !contains(integerDimensions[:], dimension.ID) || seen[dimension.ID] || dimension.Limit < 0 {
			return false
		}
		seen[dimension.ID] = true
	}
	for _, dimension := range integerDimensions {
		if !seen[dimension] {
			return false
		}
	}
	return true
}

func validDenominator(value Denominator) bool {
	return value.Schema == DenominatorSchema && value.Version != "" && value.Cases == ExpectedCaseTotal &&
		value.IntegerObservations == ExpectedCaseTotal*IntegerDimensions && value.Claims == ExpectedCaseTotal &&
		value.Interventions == ExpectedInterventions && value.AuthorityObservations == 1
}

func expectedMinimum() IntegerSet {
	return IntegerSet{InterpretationCandidates: 1, EvidencePaths: 1}
}

func exceeds(value, budget IntegerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates ||
		value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func derivedClass(program ProgramObservation, budget IntegerSet) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	if program.Counts == expectedMinimum() {
		return "ZERO"
	}
	if program.Counts == budget {
		return "BOUNDARY"
	}
	if exceeds(program.Counts, budget) {
		return "OVER"
	}
	return "WITHIN"
}

func inputState(program ProgramObservation) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	return "KNOWN"
}

func claimTarget(decision string) string {
	switch decision {
	case "PASS":
		return "DISCHARGED"
	case "FAIL_CLOSED":
		return "REFUTED"
	default:
		return "OPEN"
	}
}

func proposition(caseID string, policy BudgetPolicy) string {
	return "counts-within-budget(case:" + caseID + ",budget:" + policy.ID + ")"
}

func subjectDecision(program ProgramObservation, budget IntegerSet) (string, string, string) {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COORDINATE_UNOBSERVED"
	}
	if program.Counts.InterpretationCandidates < 1 || program.Counts.EvidencePaths < 1 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COUNT_UNKNOWN"
	}
	if exceeds(program.Counts, budget) {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"
	}
	return "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"
}
