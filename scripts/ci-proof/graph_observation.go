package main

import "fmt"

// This describes observer coverage, never the existence of a language capability.
// The deferred command is a next operation, not an invocation receipt.
type graphObservation struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func validateGraphObservation(graph domainCommand) error {
	observation := graph.Observation
	if observation == nil {
		// Historical v2 receipts predate observer coverage. Retain their bytes;
		// missing coverage is not a capability result or a new execution claim.
		return nil
	}
	if graph.Command != "go run ./cmd/gooo graph dump examples/billing/main.gooo" || graph.Status != "deferred" || graph.Available {
		return fmt.Errorf("graph observer coverage must describe the unexecuted graph dump operation")
	}
	if observation.State != "UNKNOWN" || observation.Stage != "DOMAIN_EVIDENCE" || observation.Step != "GRAPH_DUMP" {
		return fmt.Errorf("graph observer UNKNOWN stage or step is missing or invalid")
	}
	if observation.Reason != "GRAPH_OBSERVER_NOT_RUN" || graph.Reason != observation.Reason || observation.UnknownClass != "DIRECT_MISSING" {
		return fmt.Errorf("graph observer coverage must not assert capability absence")
	}
	if observation.NextOperation != "RUN_GRAPH_DUMP_OBSERVER" || observation.BlockedBy == nil || len(observation.BlockedBy) != 0 {
		return fmt.Errorf("unexecuted graph observer requires its next operation and an empty direct-missing frontier")
	}
	// An empty digest is a legacy envelope placeholder, not observed output.
	if graph.Output != "" || graph.OutputSHA256 != digestBytes(nil) {
		return fmt.Errorf("unexecuted graph observer must not carry fabricated output")
	}
	return nil
}
