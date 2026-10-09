package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

// Plain input-only execution returns one entry value per input row. --json
// retains the complete generation, execution and unobserved-expectation record.
func writePackageActualValues(stdout, stderr io.Writer, result workspaceexecution.Result) int {
	return writePackageEntryValues(stdout, stderr, result.Program.Entry.LoweredName, result.Composition.Plan, result.Runtime)
}

func writePackageEntryValues(stdout, stderr io.Writer, entryName string,
	plan bodyexecution.CompositionPlan, runtime bodyexecution.CompositionRuntime) int {
	entryID := ""
	for _, activity := range plan.Activities {
		if activity.Name == entryName {
			entryID = activity.ID
		}
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	for _, trace := range runtime.Traces {
		found := false
		for _, delivery := range trace.Deliveries {
			if entryID == "" || delivery.ActivityID != entryID {
				continue
			}
			if delivery.Fault != nil || len(delivery.BlockedBy) != 0 {
				kind := ""
				if delivery.Fault != nil {
					kind = delivery.Fault.Kind
				}
				fmt.Fprintf(stderr, "gooo package: input row %d has no entry value; fault=%s blocked_by=%v; use --json for the execution record\n",
					trace.CaseIndex, kind, delivery.BlockedBy)
				return exitFailure
			}
			if err := encoder.Encode(delivery.Actual); err != nil {
				fmt.Fprintf(stderr, "gooo package: write actual value: %v\n", err)
				return exitFailure
			}
			found = true
		}
		if !found {
			fmt.Fprintln(stderr, "gooo package: entry output was not observed")
			return exitFailure
		}
	}
	return exitOK
}
