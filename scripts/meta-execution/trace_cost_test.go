package main

import (
	"testing"
	"time"
)

func TestCostSeparatesPassesAndRetainsFailureDuration(t *testing.T) {
	var state metaExecutionCostState
	start := time.Unix(0, 0)
	event := metaExecutionTraceEvent{Boundary: "PROCESS_CALL_ENTERED", OperationSequence: 1,
		Pass: "first", CommandKind: "verifier", EventSequence: 1}
	state.observe(event, start)
	event.Pass, event.EventSequence = "replay", 2
	state.observe(event, start.Add(time.Second))
	event.Boundary, event.Pass = "PROCESS_RETURNED", "first"
	failed, code := true, 1
	event.ReturnErrorObserved, event.ExitCode = &failed, &code
	got := state.observe(event, start.Add(3*time.Second))
	if got.State != "OBSERVED" || got.ElapsedNS == nil || *got.ElapsedNS != 3000000000 || got.StartedAtEvent != 1 {
		t.Fatalf("first pass cost: %+v", got)
	}
	if got.Improvement != "UNKNOWN" || got.ToolchainIdentity != "UNOBSERVED" {
		t.Fatalf("cost invented stronger evidence: %+v", got)
	}
	event.Pass = "replay"
	got = state.observe(event, start.Add(5*time.Second))
	if got.ElapsedNS == nil || *got.ElapsedNS != 4000000000 || got.StartedAtEvent != 2 {
		t.Fatalf("replay cost: %+v", got)
	}
}

func TestCostMissingStartAndNegativeClockStayUnknown(t *testing.T) {
	var state metaExecutionCostState
	event := metaExecutionTraceEvent{Boundary: "ACTION_RETURNED"}
	if got := state.observe(event, time.Unix(2, 0)); got.State != "UNKNOWN" || got.ElapsedNS != nil {
		t.Fatalf("unpaired return: %+v", got)
	}
	event.Boundary, event.EventSequence = "ACTION_ENTERED", 1
	state.observe(event, time.Unix(2, 0))
	event.Boundary = "ACTION_RETURNED"
	if got := state.observe(event, time.Unix(1, 0)); got.State != "UNKNOWN" || got.ElapsedNS != nil {
		t.Fatalf("negative elapsed: %+v", got)
	}
}
