package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime/workspaceexecution"
)

// Plain input-only execution returns one entry value per input row. --json
// retains the complete generation, execution and unobserved-expectation record.
func writePackageActualValues(stdout, stderr io.Writer, result workspaceexecution.Result) int {
	entryID := ""
	for _, activity := range result.Composition.Plan.Activities {
		if activity.Name == result.Program.Entry.LoweredName {
			entryID = activity.ID
		}
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	for _, trace := range result.Runtime.Traces {
		found := false
		for _, delivery := range trace.Deliveries {
			if entryID == "" || delivery.ActivityID != entryID {
				continue
			}
			if err := encoder.Encode(delivery.Actual); err != nil {
				fmt.Fprintf(stderr, "gooo package execute: write actual value: %v\n", err)
				return exitFailure
			}
			found = true
		}
		if !found {
			fmt.Fprintln(stderr, "gooo package execute: entry output was not observed")
			return exitFailure
		}
	}
	return exitOK
}
