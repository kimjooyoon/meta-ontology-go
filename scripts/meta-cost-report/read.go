package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

func readCostReport(input io.Reader) (costReport, error) {
	report := costReport{Schema: "gooo/meta-execution-cost-report/v1",
		Scope: "DIAGNOSTIC_INTERVALS_ONLY_NO_ADDITIVE_TOTAL",
		Authenticity: "UNVERIFIED", Improvement: "UNKNOWN", Rows: []costRow{}}
	events := make(map[eventKey]costEvent)
	ordered := []costEvent{}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	for scanner.Scan() {
		var event costEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return report, fmt.Errorf("invalid NDJSON event %d: %w", len(ordered)+1, err)
		}
		if event.Schema != "gooo/meta-execution-driver-boundary/v1" || event.Invocation == "" || event.Sequence == 0 {
			return report, fmt.Errorf("invalid event identity at line %d", len(ordered)+1)
		}
		key := eventKey{event.Invocation, event.Sequence}
		if _, exists := events[key]; exists {
			return report, fmt.Errorf("duplicate event: %s/%d", key.invocation, key.sequence)
		}
		events[key] = event
		ordered = append(ordered, event)
	}
	if err := scanner.Err(); err != nil {
		return report, err
	}
	return assembleCostReport(report, events, ordered)
}
