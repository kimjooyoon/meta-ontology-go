package verify

import "sort"

type violationKey struct {
	Path   string
	Rule   string
	Detail string
}

func violationKeyOf(violation Violation) violationKey {
	return violationKey{Path: violation.Path, Rule: violation.Rule, Detail: violation.Detail}
}

func policyViolationRegressed(current, previous []Violation) bool {
	if len(current) == 0 {
		return false
	}
	key := violationKeyOf(current[0])
	current = matchingViolations(current, key)
	previous = matchingViolations(previous, key)
	if len(previous) == 0 || len(current) > len(previous) {
		return true
	}
	current = sortedActuals(current)
	previous = sortedActuals(previous)
	for index, violation := range current {
		if violation.Actual > previous[index].Actual {
			return true
		}
	}
	return false
}

func matchingViolations(violations []Violation, key violationKey) []Violation {
	matching := make([]Violation, 0)
	for _, violation := range violations {
		if violationKeyOf(violation) == key {
			matching = append(matching, violation)
		}
	}
	return matching
}

func sortedActuals(violations []Violation) []Violation {
	sorted := append([]Violation(nil), violations...)
	sort.Slice(sorted, func(left, right int) bool {
		return sorted[left].Actual < sorted[right].Actual
	})
	return sorted
}
