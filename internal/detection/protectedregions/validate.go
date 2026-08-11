package protectedregions

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type sourceLine struct {
	start int
	end   int
	text  string
}

type markerEvent struct {
	kind     MarkerKind
	boundary MarkerBoundary
	id       string
	line     int
	start    int
	end      int
}

type openMarker struct {
	event     markerEvent
	bodyStart int
}

// Validate checks marker pairing, ownership nesting, and stable marker IDs.
// The accepted generated syntax is compatible with the generator contract:
//
//	//gooo:generated:start id="..." kind="..."
//gooo:generated:end id="..."
//gooo:slot:start id="..."
//gooo:slot:end id="..."

// Handwritten regions use handwritten or protected in the marker name:
//
//	//gooo:handwritten:start id="..."
//gooo:handwritten:end id="..."

func Validate(source []byte) Report {
	lines := splitSourceLines(source)
	result := Report{}
	stack := make([]openMarker, 0)
	seen := make(map[string]int)
	for index, line := range lines {
		event, ok, err := parseMarker(line, index+1)
		if err != nil {
			result.Issues = append(result.Issues, Issue{Kind: IssueInvalidMarker, Line: index + 1, Detail: err.Error()})
			continue
		}
		if !ok {
			continue
		}
		if event.id == "" {
			result.Issues = append(result.Issues, Issue{
				Kind: IssueMissingID, Marker: event.kind, Line: event.line,
				Detail: "marker has no id",
			})
		}
		if event.boundary == Start {
			key := string(event.kind) + "\x00" + event.id
			if event.id != "" {
				if previous, exists := seen[key]; exists {
					result.Issues = append(result.Issues, Issue{
						Kind: IssueDuplicateMarker, Marker: event.kind, ID: event.id, Line: event.line,
						Detail: fmt.Sprintf("also opened on line %d", previous),
					})
				}
				seen[key] = event.line
			}
			if issue := startIssue(event, stack); issue != nil {
				result.Issues = append(result.Issues, *issue)
			}
			stack = append(stack, openMarker{event: event, bodyStart: event.end})
			continue
		}
		var opened openMarker
		var closed bool
		stack, opened, closed = closeMarker(event, stack, &result)
		if !closed {
			continue
		}
		result.Regions = append(result.Regions, Region{
			Kind:      opened.event.kind,
			ID:        opened.event.id,
			StartLine: opened.event.line,
			EndLine:   event.line,
			Start:     opened.event.start,
			End:       event.end,
			BodyStart: opened.bodyStart,
			BodyEnd:   event.start,
		})
	}
	for index := len(stack) - 1; index >= 0; index-- {
		opened := stack[index].event
		result.Issues = append(result.Issues, Issue{
			Kind: IssueMissingEnd, Marker: opened.kind, ID: opened.id, Line: opened.line,
			Detail: "marker is not closed",
		})
	}
	sort.SliceStable(result.Regions, func(i, j int) bool { return result.Regions[i].Start < result.Regions[j].Start })
	return result
}

func startIssue(event markerEvent, stack []openMarker) *Issue {
	if event.kind == Generated && len(stack) > 0 {
		return &Issue{
			Kind: IssueNestedMarker, Marker: event.kind, ID: event.id, Line: event.line,
			Detail: "generated region is nested inside another region",
		}
	}
	if event.kind == Slot {
		if !hasOpenKind(stack, Generated) {
			return &Issue{
				Kind: IssueSlotOutsideGenerated, Marker: event.kind, ID: event.id, Line: event.line,
				Detail: "slot must be inside a generated region",
			}
		}
		if topKind(stack, Slot) || topKind(stack, Handwritten) {
			return &Issue{
				Kind: IssueNestedMarker, Marker: event.kind, ID: event.id, Line: event.line,
				Detail: "slot is nested inside another protected region",
			}
		}
	}
	if event.kind == Handwritten && (topKind(stack, Slot) || topKind(stack, Handwritten)) {
		return &Issue{
			Kind: IssueNestedMarker, Marker: event.kind, ID: event.id, Line: event.line,
			Detail: "handwritten region is nested inside another protected region",
		}
	}
	return nil
}

func closeMarker(event markerEvent, stack []openMarker, result *Report) ([]openMarker, openMarker, bool) {
	if len(stack) == 0 {
		result.Issues = append(result.Issues, Issue{Kind: IssueMissingStart, Marker: event.kind, ID: event.id, Line: event.line, Detail: "closing marker has no matching start"})
		return stack, openMarker{}, false
	}
	top := stack[len(stack)-1]
	if top.event.kind != event.kind {
		if hasOpenKind(stack, event.kind) {
			result.Issues = append(result.Issues, Issue{Kind: IssueNestedMarker, Marker: event.kind, ID: event.id, Line: event.line, Detail: fmt.Sprintf("closing marker crosses open %s region", top.event.kind)})
		} else {
			result.Issues = append(result.Issues, Issue{Kind: IssueMissingStart, Marker: event.kind, ID: event.id, Line: event.line, Detail: "closing marker has no matching start"})
		}
		return stack, openMarker{}, false
	}
	if event.id != "" && top.event.id != "" && event.id != top.event.id {
		result.Issues = append(result.Issues, Issue{Kind: IssueMismatchedMarker, Marker: event.kind, ID: event.id, Line: event.line, Detail: fmt.Sprintf("opened as %q", top.event.id)})
	}
	return stack[:len(stack)-1], top, true
}

func hasOpenKind(stack []openMarker, kind MarkerKind) bool {
	for _, open := range stack {
		if open.event.kind == kind {
			return true
		}
	}
	return false
}

func topKind(stack []openMarker, kind MarkerKind) bool {
	return len(stack) > 0 && stack[len(stack)-1].event.kind == kind
}

func parseMarker(line sourceLine, lineNumber int) (markerEvent, bool, error) {
	trimmed := strings.TrimSpace(line.text)
	for _, spec := range markerSpecs() {
		if !hasMarkerPrefix(trimmed, spec.prefix) {
			continue
		}
		rest := strings.TrimSpace(trimmed[len(spec.prefix):])
		id, err := markerID(rest, spec.legacy)
		if err != nil {
			return markerEvent{}, true, err
		}
		return markerEvent{kind: spec.kind, boundary: spec.boundary, id: id, line: lineNumber, start: line.start, end: line.end}, true, nil
	}
	return markerEvent{}, false, nil
}

func splitSourceLines(source []byte) []sourceLine {
	lines := make([]sourceLine, 0)
	for start := 0; start < len(source); {
		relativeEnd := bytes.IndexByte(source[start:], '\n')
		end := len(source)
		if relativeEnd >= 0 {
			end = start + relativeEnd + 1
		}
		textEnd := end
		if textEnd > start && source[textEnd-1] == '\n' {
			textEnd--
		}
		if textEnd > start && source[textEnd-1] == '\r' {
			textEnd--
		}
		lines = append(lines, sourceLine{start: start, end: end, text: string(source[start:textEnd])})
		start = end
	}
	return lines
}
