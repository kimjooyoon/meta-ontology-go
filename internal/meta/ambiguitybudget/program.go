package ambiguitybudget

import (
	"fmt"
	"strconv"
	"strings"
)

const programPrefix = "ambiguity-budget"

type computesProgram struct {
	Activity   string
	Text       string
	Kind       string
	ID         string
	Class      string
	InputState string
	Counts     IntegerSet
}

func expectedBudget() IntegerSet {
	return IntegerSet{InterpretationCandidates: 2, UnresolvedBranches: 1, EvidencePaths: 2}
}

func parseComputesProgram(activity, text string) (computesProgram, error) {
	if activity == "" || text == "" || text != strings.TrimSpace(text) || strings.ContainsAny(text, " \t\r\n") {
		return computesProgram{}, fmt.Errorf("non-canonical computes program")
	}
	parts := strings.Split(text, ":")
	if len(parts) == 3 && parts[0] == programPrefix && parts[1] == "budget" {
		counts, err := parseIntegerSet(parts[2])
		if err != nil {
			return computesProgram{}, err
		}
		return computesProgram{Activity: activity, Text: text, Kind: "BUDGET", ID: "budget", Counts: counts}, nil
	}
	if len(parts) != 6 || parts[0] != programPrefix || parts[1] != "case" || parts[2] == "" {
		return computesProgram{}, fmt.Errorf("unsupported computes program %q", text)
	}
	counts, err := parseIntegerSet(parts[5])
	if err != nil {
		return computesProgram{}, err
	}
	if parts[3] != "ZERO" && parts[3] != "BOUNDARY" && parts[3] != "OVER" && parts[3] != "UNKNOWN" {
		return computesProgram{}, fmt.Errorf("unknown ambiguity case class %q", parts[3])
	}
	if parts[4] != "KNOWN" && parts[4] != "UNKNOWN" {
		return computesProgram{}, fmt.Errorf("unknown ambiguity input state %q", parts[4])
	}
	return computesProgram{Activity: activity, Text: text, Kind: "CASE", ID: parts[2], Class: parts[3], InputState: parts[4], Counts: counts}, nil
}

func parseIntegerSet(text string) (IntegerSet, error) {
	parts := strings.Split(text, ",")
	if len(parts) != IntegerDimensions {
		return IntegerSet{}, fmt.Errorf("integer set must have %d coordinates", IntegerDimensions)
	}
	values := [IntegerDimensions]int{}
	for index, part := range parts {
		if part == "" || part != strings.TrimSpace(part) {
			return IntegerSet{}, fmt.Errorf("integer set coordinate is not canonical")
		}
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return IntegerSet{}, fmt.Errorf("integer set coordinate %q is not a non-negative integer", part)
		}
		values[index] = value
	}
	return IntegerSet{InterpretationCandidates: values[0], UnresolvedBranches: values[1], EvidencePaths: values[2]}, nil
}

func formatIntegerSet(value IntegerSet) string {
	return fmt.Sprintf("%d,%d,%d", value.InterpretationCandidates, value.UnresolvedBranches, value.EvidencePaths)
}

func formatComputesProgram(program computesProgram) string {
	if program.Kind == "BUDGET" {
		return programPrefix + ":budget:" + formatIntegerSet(program.Counts)
	}
	return strings.Join([]string{programPrefix, "case", program.ID, program.Class, program.InputState, formatIntegerSet(program.Counts)}, ":")
}

func validCounts(counts IntegerSet) bool {
	return counts.InterpretationCandidates >= 1 && counts.UnresolvedBranches >= 0 && counts.EvidencePaths >= 1
}

func exceeds(value, budget IntegerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates ||
		value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func subjectDecision(program computesProgram, budget IntegerSet) (string, string, string, string) {
	if program.InputState == "UNKNOWN" {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_INPUT_UNKNOWN", "OPEN"
	}
	if !validCounts(program.Counts) {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COUNT_UNKNOWN", "OPEN"
	}
	if exceeds(program.Counts, budget) {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED", "LOWER_RESOLUTION"
	}
	return "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT", "EXACT"
}
