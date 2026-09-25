package roundtrip

import (
	"fmt"
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
