package protectedregions

import (
	"fmt"
)

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
	if event.kind == Generated && !event.legacy && !top.event.legacy &&
		event.semanticKind != "" && top.event.semanticKind != "" &&
		event.semanticKind != top.event.semanticKind {
		result.Issues = append(result.Issues, Issue{
			Kind: IssueMismatchedMarker, Marker: event.kind, ID: event.id, Line: event.line,
			Detail: fmt.Sprintf("opened with kind %q", top.event.semanticKind),
		})
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
