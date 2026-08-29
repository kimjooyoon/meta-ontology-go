package verify

import (
	"sort"
	"strings"
)

func policyRegressions(root, base string, current []Violation, policy LinePolicy) ([]Violation, error) {
	previous := make(map[string][]Violation)
	loaded := make(map[string]bool)
	regressions := make([]Violation, 0)
	groups := make(map[violationKey][]Violation)
	for _, violation := range current {
		if !isCapViolation(violation) {
			regressions = append(regressions, violation)
			continue
		}
		key := violationKeyOf(violation)
		groups[key] = append(groups[key], violation)
	}
	for key, group := range groups {
		if !loaded[key.Path] {
			var err error
			previous[key.Path], err = revisionViolations(root, base, key.Path, policy)
			if err != nil {
				return nil, err
			}
			loaded[key.Path] = true
		}
		if policyViolationRegressed(group, previous[key.Path]) {
			regressions = append(regressions, group...)
		}
	}
	sort.Slice(regressions, func(left, right int) bool { return regressions[left].Error() < regressions[right].Error() })
	return regressions, nil
}

func isCapViolation(violation Violation) bool {
	return violation.Rule == "DAMP file lines" || violation.Rule == "DRY function lines" || violation.Rule == "GOOO file lines"
}

func formatViolations(violations []Violation) string {
	lines := make([]string, len(violations))
	for index, violation := range violations {
		lines[index] = violation.Error()
	}
	return strings.Join(lines, "\n")
}
