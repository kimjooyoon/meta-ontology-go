package protectedregions

import (
	"fmt"
	"strings"
)

// Body returns the bytes between the marker lines. It is safe for a Region
// returned by Validate to be used with the source passed to Validate.
func (r Region) Body(source []byte) []byte {
	if r.BodyStart < 0 || r.BodyEnd < r.BodyStart || r.BodyEnd > len(source) {
		return nil
	}
	return source[r.BodyStart:r.BodyEnd]
}

// Report is the result of structural validation. Regions are returned in
// source order and Issues are returned in deterministic source order.
type Report struct {
	Regions []Region
	Issues  []Issue
}

// Valid reports whether all marker and nesting checks passed.
func (r Report) Valid() bool { return len(r.Issues) == 0 }

// Err turns a report into a single deterministic error for callers that use a
// fail-fast validation boundary.
func (r Report) Err() error {
	if r.Valid() {
		return nil
	}
	lines := make([]string, len(r.Issues))
	for index, issue := range r.Issues {
		lines[index] = issue.Error()
	}
	return fmt.Errorf("protected-region validation failed:\n%s", strings.Join(lines, "\n"))
}

// LocalityIssueKind classifies a change that escapes its generated boundary.
type LocalityIssueKind string

const (
	LocalityUnownedChange   LocalityIssueKind = "unowned-change"
	LocalityProtectedChange LocalityIssueKind = "protected-change"
)

// LocalityIssue describes a before/after change that a safe generator merge
// must reject.
type LocalityIssue struct {
	Kind   LocalityIssueKind
	Marker MarkerKind
	ID     string
	Detail string
}

func (i LocalityIssue) Error() string {
	if i.Detail == "" {
		return fmt.Sprintf("%s: %s %q", i.Kind, i.Marker, i.ID)
	}
	return fmt.Sprintf("%s: %s %q: %s", i.Kind, i.Marker, i.ID, i.Detail)
}

// LocalityReport contains structural reports and locality violations for a
// source refresh. Generated bodies may change; protected bodies and all text
// outside generated regions may not.
type LocalityReport struct {
	Before     Report
	After      Report
	Violations []LocalityIssue
}
