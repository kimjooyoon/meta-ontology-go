package ambiguitybudget

import (
	"fmt"
	"strconv"
	"strings"
)

const programPrefix = "ambiguity-budget"

type computesProgram struct {
	Activity             string
	Text                 string
	Kind                 string
	ID                   string
	Counts               IntegerSet
	UnobservedDimensions []string
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
		counts, unobserved, err := parseIntegerSet(parts[2])
		if err != nil {
			return computesProgram{}, err
		}
		if len(unobserved) != 0 {
			return computesProgram{}, fmt.Errorf("budget coordinates must be observed")
		}
		return computesProgram{Activity: activity, Text: text, Kind: "BUDGET", ID: "budget", Counts: counts}, nil
	}
	if len(parts) != 4 || parts[0] != programPrefix || parts[1] != "case" || parts[2] == "" {
		return computesProgram{}, fmt.Errorf("unsupported computes program %q", text)
	}
	counts, unobserved, err := parseIntegerSet(parts[3])
	if err != nil {
		return computesProgram{}, err
	}
	return computesProgram{Activity: activity, Text: text, Kind: "CASE", ID: parts[2], Counts: counts, UnobservedDimensions: unobserved}, nil
}

func parseIntegerSet(text string) (IntegerSet, []string, error) {
	parts := strings.Split(text, ",")
	if len(parts) != IntegerDimensions {
		return IntegerSet{}, nil, fmt.Errorf("integer set must have %d coordinates", IntegerDimensions)
	}
	values := [IntegerDimensions]int{}
	unobserved := make([]string, 0, IntegerDimensions)
	for index, part := range parts {
		if part == "" || part != strings.TrimSpace(part) {
			return IntegerSet{}, nil, fmt.Errorf("integer set coordinate is not canonical")
		}
		if part == "?" {
			unobserved = append(unobserved, integerDimensions[index])
			continue
		}
		value, err := strconv.Atoi(part)
		if err != nil || value < 0 {
			return IntegerSet{}, nil, fmt.Errorf("integer set coordinate %q is not a non-negative integer", part)
		}
		values[index] = value
	}
	return IntegerSet{InterpretationCandidates: values[0], UnresolvedBranches: values[1], EvidencePaths: values[2]}, unobserved, nil
}

var integerDimensions = [...]string{"interpretation_candidates", "unresolved_branches", "evidence_paths"}

func formatIntegerSet(value IntegerSet, unobserved []string) string {
	values := []string{strconv.Itoa(value.InterpretationCandidates), strconv.Itoa(value.UnresolvedBranches), strconv.Itoa(value.EvidencePaths)}
	for index, dimension := range integerDimensions {
		for _, missing := range unobserved {
			if missing == dimension {
				values[index] = "?"
			}
		}
	}
	return strings.Join(values, ",")
}

func formatComputesProgram(program computesProgram) string {
	if program.Kind == "BUDGET" {
		return programPrefix + ":budget:" + formatIntegerSet(program.Counts, nil)
	}
	return strings.Join([]string{programPrefix, "case", program.ID, formatIntegerSet(program.Counts, program.UnobservedDimensions)}, ":")
}

func validCounts(counts IntegerSet, unobserved []string) bool {
	if !validUnobserved(unobserved) {
		return false
	}
	if contains(unobserved, "interpretation_candidates") || contains(unobserved, "evidence_paths") {
		return counts.InterpretationCandidates >= 0 && counts.UnresolvedBranches >= 0 && counts.EvidencePaths >= 0
	}
	return counts.InterpretationCandidates >= 1 && counts.UnresolvedBranches >= 0 && counts.EvidencePaths >= 1
}

func validUnobserved(unobserved []string) bool {
	if len(unobserved) == 0 {
		return true
	}
	seen := make(map[string]bool, len(unobserved))
	for _, dimension := range unobserved {
		if !contains(integerDimensions[:], dimension) || seen[dimension] {
			return false
		}
		seen[dimension] = true
	}
	return true
}

func exceeds(value IntegerSet, budget IntegerSet) bool {
	return value.InterpretationCandidates > budget.InterpretationCandidates ||
		value.UnresolvedBranches > budget.UnresolvedBranches || value.EvidencePaths > budget.EvidencePaths
}

func derivedClass(program computesProgram, budget IntegerSet) string {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN"
	}
	if program.Counts == (IntegerSet{InterpretationCandidates: 1, UnresolvedBranches: 0, EvidencePaths: 1}) {
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

func inputState(program computesProgram) string {
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

func subjectDecision(program computesProgram, budget IntegerSet) (string, string, string) {
	if len(program.UnobservedDimensions) > 0 {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COORDINATE_UNOBSERVED"
	}
	if !validCounts(program.Counts, nil) {
		return "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COUNT_UNKNOWN"
	}
	if exceeds(program.Counts, budget) {
		return "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"
	}
	return "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"
}
