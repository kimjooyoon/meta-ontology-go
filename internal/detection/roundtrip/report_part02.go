package roundtrip

import (
	"sort"
)

func (r *Report) merge(other Report) {
	if r == nil {
		return
	}
	r.add(other.Violations...)
}
func (r *Report) normalize() {
	if r == nil {
		return
	}
	sort.SliceStable(r.Violations, func(i, j int) bool {
		left, right := r.Violations[i], r.Violations[j]
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Identity != right.Identity {
			return left.Identity < right.Identity
		}
		if left.Expected != right.Expected {
			return left.Expected < right.Expected
		}
		if left.Actual != right.Actual {
			return left.Actual < right.Actual
		}
		return left.Detail < right.Detail
	})
}
func snapshotViolation(path string, err error) Violation {
	return Violation{Rule: RuleSnapshot, Path: path, Detail: err.Error()}
}
func semanticViolation(rule, identity, expected, actual, detail string) Violation {
	return Violation{Rule: rule, Path: "semantic", Identity: identity, Expected: expected, Actual: actual, Detail: detail}
}
