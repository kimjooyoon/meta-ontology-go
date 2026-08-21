package protectedregions

import (
	"fmt"
	"sort"
)

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
		if event.id == "" && !(event.legacy && event.boundary == End) {
			result.Issues = append(result.Issues, Issue{Kind: IssueMissingID, Marker: event.kind, Line: event.line, Detail: "marker has no id"})
		}
		if event.kind == Generated && !event.legacy && event.semanticKind == "" {
			result.Issues = append(result.Issues, Issue{
				Kind: IssueMissingKind, Marker: event.kind, ID: event.id, Line: event.line,
				Detail: "generated marker has no kind",
			})
		}
		if event.boundary == Start {
			key := event.id
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
	for index := 0; index < len(stack); index++ {
		opened := stack[index].event
		result.Issues = append(result.Issues, Issue{
			Kind: IssueMissingEnd, Marker: opened.kind, ID: opened.id, Line: opened.line,
			Detail: "marker is not closed",
		})
	}
	sort.SliceStable(result.Issues, func(i, j int) bool { return result.Issues[i].Line < result.Issues[j].Line })
	sort.SliceStable(result.Regions, func(i, j int) bool { return result.Regions[i].Start < result.Regions[j].Start })
	return result
}
