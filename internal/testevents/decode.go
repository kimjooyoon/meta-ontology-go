// Package testevents decodes named Go test observations without choosing their
// expected outcome. Package-level events remain outside this named-test view.
package testevents

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Event struct {
	Name   string `json:"name"`
	Action string `json:"action"`
}

// Decode retains source order and a decoded failure prefix when a later event
// is malformed or contradictory. Successful decoding does not require a pass.
func Decode(raw []byte) ([]Event, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	events := []Event{}
	seen := map[string]bool{}
	for {
		var event struct{ Action, Test string }
		if err := decoder.Decode(&event); err == io.EOF {
			return events, nil
		} else if err != nil {
			return events, fmt.Errorf("package test event decoding: %w", err)
		}
		if event.Test == "" {
			continue
		}
		switch event.Action {
		case "run", "output", "pause", "cont", "bench":
			continue
		case "pass", "fail", "skip":
		default:
			return events, fmt.Errorf("unknown package test action %q", event.Action)
		}
		if seen[event.Test] {
			if event.Action == "fail" {
				events = append(events, Event{Name: event.Test, Action: event.Action})
			}
			return events, fmt.Errorf("duplicate terminal test event for %s", event.Test)
		}
		seen[event.Test] = true
		events = append(events, Event{Name: event.Test, Action: event.Action})
	}
}
