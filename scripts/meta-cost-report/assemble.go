package main

import (
	"fmt"
	"sort"
)

func assembleCostReport(report costReport, events map[eventKey]costEvent, ordered []costEvent) (costReport, error) {
	starts, used := make(map[eventKey]bool), make(map[eventKey]bool)
	report.Events = len(ordered)
	for _, event := range ordered {
		if event.Cost == nil {
			report.UnmeasuredEvents++
			continue
		}
		key := eventKey{event.Invocation, event.Sequence}
		switch event.Cost.State {
		case "STARTED":
			if event.Boundary != "ACTION_ENTERED" && event.Boundary != "PROCESS_CALL_ENTERED" {
				return report, fmt.Errorf("STARTED at non-entry boundary: %v", key)
			}
			starts[key] = true
		case "UNKNOWN":
			report.UnknownReturns++
		case "OBSERVED":
			startKey := eventKey{event.Invocation, event.Cost.Start}
			start, exists := events[startKey]
			if !exists || used[startKey] || !validCostPair(start, event) {
				return report, fmt.Errorf("invalid or reused interval start: %v", key)
			}
			used[startKey] = true
			report.Rows = append(report.Rows, costRow{event.costBinding, event.Cost.Start, event.Sequence, *event.Cost.Elapsed})
		default:
			return report, fmt.Errorf("unknown cost decision %q", event.Cost.State)
		}
	}
	for key := range starts {
		if !used[key] {
			report.UnpairedStarts++
		}
	}
	sort.Slice(report.Rows, func(i, j int) bool {
		left, right := report.Rows[i], report.Rows[j]
		if left.Invocation != right.Invocation {
			return left.Invocation < right.Invocation
		}
		return left.Return < right.Return
	})
	return report, nil
}

func validCostPair(start, end costEvent) bool {
	if start.Cost == nil || start.Cost.State != "STARTED" || start.Sequence >= end.Sequence ||
		start.costBinding != end.costBinding || end.Cost.Elapsed == nil || *end.Cost.Elapsed < 0 {
		return false
	}
	return start.Boundary == "ACTION_ENTERED" && end.Boundary == "ACTION_RETURNED" ||
		start.Boundary == "PROCESS_CALL_ENTERED" && end.Boundary == "PROCESS_RETURNED"
}
