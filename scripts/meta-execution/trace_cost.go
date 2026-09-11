package main

import "time"

// Cost is diagnostic, never part of canonical output or permission evidence.
type metaExecutionCost struct {
	State             string `json:"state"`
	StartedAtEvent    uint64 `json:"started_at_event,omitempty"`
	ElapsedNS         *int64 `json:"elapsed_ns,omitempty"`
	ExecutionMode     string `json:"execution_mode"`
	ToolchainIdentity string `json:"toolchain_identity"`
	Improvement       string `json:"improvement"`
}

type metaExecutionCostKey struct {
	sequence int
	pass     string
	kind     string
	family   string
}

type metaExecutionCostStart struct {
	at    time.Time
	event uint64
}

type metaExecutionCostState struct {
	starts map[metaExecutionCostKey][]metaExecutionCostStart
}

func (state *metaExecutionCostState) observe(event metaExecutionTraceEvent, now time.Time) *metaExecutionCost {
	family, entered := "", false
	switch event.Boundary {
	case "ACTION_ENTERED":
		family, entered = "action", true
	case "ACTION_RETURNED":
		family = "action"
	case "PROCESS_CALL_ENTERED":
		family, entered = "process", true
	case "PROCESS_RETURNED":
		family = "process"
	default:
		return nil
	}
	key := metaExecutionCostKey{event.OperationSequence, event.Pass, event.CommandKind, family}
	if state.starts == nil {
		state.starts = make(map[metaExecutionCostKey][]metaExecutionCostStart)
	}
	cost := &metaExecutionCost{State: "UNKNOWN", ExecutionMode: "OBSERVED_DRIVER_CALL",
		ToolchainIdentity: "UNOBSERVED", Improvement: "UNKNOWN"}
	if entered {
		state.starts[key] = append(state.starts[key], metaExecutionCostStart{now, event.EventSequence})
		cost.State = "STARTED"
		return cost
	}
	stack := state.starts[key]
	if len(stack) == 0 {
		return cost
	}
	start := stack[len(stack)-1]
	state.starts[key] = stack[:len(stack)-1]
	cost.StartedAtEvent = start.event
	elapsed := now.Sub(start.at).Nanoseconds()
	if elapsed >= 0 {
		cost.State, cost.ElapsedNS = "OBSERVED", &elapsed
	}
	return cost
}
