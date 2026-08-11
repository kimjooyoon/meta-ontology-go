// Package protectedregions validates the ownership boundaries in generated Go.
//
// A generated region may contain handwritten slots.  Handwritten regions are
// protected from replacement, while text outside generated regions must also
// remain local when a projection is refreshed.
package protectedregions

import (
	"fmt"
	"strings"
)

// MarkerKind identifies the kind of owned region a marker opens or closes.
type MarkerKind string

const (
	Generated   MarkerKind = "generated"
	Slot        MarkerKind = "slot"
	Handwritten MarkerKind = "handwritten"
)

// MarkerBoundary identifies whether a marker opens or closes a region.
type MarkerBoundary string

const (
	Start MarkerBoundary = "start"
	End   MarkerBoundary = "end"
)

// IssueKind classifies a malformed ownership boundary.
type IssueKind string

const (
	IssueInvalidMarker        IssueKind = "invalid-marker"
	IssueMissingID            IssueKind = "missing-id"
	IssueMissingStart         IssueKind = "missing-start"
	IssueMissingEnd           IssueKind = "missing-end"
	IssueDuplicateMarker      IssueKind = "duplicate-marker"
	IssueMismatchedMarker     IssueKind = "mismatched-marker"
	IssueNestedMarker         IssueKind = "nested-marker"
	IssueSlotOutsideGenerated IssueKind = "slot-outside-generated"
)

// Issue describes one structural risk. Line is one-based; a missing-end issue
// points to the opening line because the source has no closing line.
type Issue struct {
	Kind   IssueKind
	Marker MarkerKind
	ID     string
	Line   int
	Detail string
}

func (i Issue) Error() string {
	location := ""
	if i.Line > 0 {
		location = fmt.Sprintf("line %d: ", i.Line)
	}
	if i.Detail == "" {
		return fmt.Sprintf("%s%s", location, i.Kind)
	}
	return fmt.Sprintf("%s%s: %s", location, i.Kind, i.Detail)
}

// Region is a paired marker range. Offsets are byte offsets into the source;
// Body excludes the opening and closing marker lines.
type Region struct {
	Kind      MarkerKind
	ID        string
	StartLine int
	EndLine   int
	Start     int
	End       int
	BodyStart int
	BodyEnd   int
}

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

// Valid reports whether both sources are structurally sound and the change is
// confined to generated region bodies.
func (r LocalityReport) Valid() bool {
	return r.Before.Valid() && r.After.Valid() && len(r.Violations) == 0
}

// Err turns a locality report into a fail-fast error.
func (r LocalityReport) Err() error {
	if err := r.Before.Err(); err != nil {
		return err
	}
	if err := r.After.Err(); err != nil {
		return err
	}
	if len(r.Violations) == 0 {
		return nil
	}
	lines := make([]string, len(r.Violations))
	for index, violation := range r.Violations {
		lines[index] = violation.Error()
	}
	return fmt.Errorf("protected-region locality validation failed:\n%s", strings.Join(lines, "\n"))
}
