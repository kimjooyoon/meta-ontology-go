package roundtrip

import (
	"fmt"
	"sort"
	"strings"
)

const (
	RuleDSLToIR   = "dsl-to-ir"
	RuleGoToIR    = "go-to-ir"
	RuleRoundTrip = "round-trip"
	RuleLocality  = "locality"
	RuleSnapshot  = "snapshot"
	RuleMarker    = "generated-marker"
)

// Violation is one deterministic, source-independent detector finding.
type Violation struct {
	Rule     string
	Path     string
	Identity string
	Expected string
	Actual   string
	Detail   string
}

func (v Violation) Error() string {
	location := v.Path
	if v.Identity != "" {
		location += "[" + v.Identity + "]"
	}
	message := v.Detail
	if message == "" {
		message = fmt.Sprintf("expected %q, got %q", v.Expected, v.Actual)
	}
	if location == "" {
		return v.Rule + ": " + message
	}
	return location + ": " + v.Rule + ": " + message
}

// Report is the complete result of one or more detector checks.
type Report struct {
	Violations []Violation
}

// OK reports whether every requested invariant passed.
func (r Report) OK() bool { return len(r.Violations) == 0 }

// Err adapts a report to APIs that use error for a failed verification.
func (r Report) Err() error {
	if r.OK() {
		return nil
	}
	return r
}

// Error formats all findings in deterministic order.
func (r Report) Error() string {
	if r.OK() {
		return "roundtrip verification passed"
	}
	lines := make([]string, len(r.Violations))
	for index, violation := range r.Violations {
		lines[index] = violation.Error()
	}
	return "roundtrip verification failed:\n" + strings.Join(lines, "\n")
}

func (r *Report) add(violations ...Violation) {
	if r == nil {
		return
	}
	r.Violations = append(r.Violations, violations...)
}

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
		return left.Detail < right.Detail
	})
}

func snapshotViolation(path string, err error) Violation {
	return Violation{Rule: RuleSnapshot, Path: path, Detail: err.Error()}
}

func semanticViolation(rule, identity, expected, actual, detail string) Violation {
	return Violation{Rule: rule, Path: "semantic", Identity: identity, Expected: expected, Actual: actual, Detail: detail}
}
