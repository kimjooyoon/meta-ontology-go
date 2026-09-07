package main

import "sort"

// Missing terminal evidence identifies an observation gap, not its external
// cause. Do not infer cancellation, process death, or permission from this record.
type boundaryUnknown struct {
	costBinding
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	StartEvent    uint64   `json:"started_at_event"`
}

func missingBoundaryReturn(event costEvent) boundaryUnknown {
	return boundaryUnknown{
		costBinding: event.costBinding,
		State: "UNKNOWN", Stage: "DRIVER_BOUNDARY", Step: event.Boundary,
		Reason: "MATCHING_TERMINAL_OBSERVATION_MISSING", UnknownClass: "DIRECT_MISSING",
		NextOperation: "OBSERVE_MATCHING_TERMINAL_EVENT", BlockedBy: []string{},
		StartEvent: event.Sequence,
	}
}

func sortBoundaryUnknowns(unknowns []boundaryUnknown) {
	sort.Slice(unknowns, func(i, j int) bool {
		if unknowns[i].Invocation != unknowns[j].Invocation {
			return unknowns[i].Invocation < unknowns[j].Invocation
		}
		return unknowns[i].StartEvent < unknowns[j].StartEvent
	})
}
