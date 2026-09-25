package protectedregions

import (
	"fmt"
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
	IssueMissingKind          IssueKind = "missing-kind"
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
