package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

func runSourceContinuation(options runSourceOptions, plan valueexecution.Plan, input int64, runtimePlanDigest string, jsonMode bool, stdout, stderr io.Writer) int {
	trace, err := plan.ExecuteIterations(context.Background(), map[string]int64{options.entry: input}, options.iterations)
	if runtimePlanDigest != "" {
		var bindErr error
		trace, bindErr = trace.BindRuntimePlanDigest(runtimePlanDigest)
		if err == nil {
			err = bindErr
		}
	}
	decision, code := "PASS", exitOK
	if err != nil {
		decision, code = "FAIL_CLOSED", exitFailure
	}
	if jsonMode {
		payload := struct {
			Schema              string                      `json:"schema"`
			Decision            string                      `json:"decision"`
			SourcePath          string                      `json:"source_path"`
			SourceDigest        string                      `json:"source_digest"`
			SemanticFingerprint string                      `json:"semantic_fingerprint"`
			RuntimePlanDigest   string                      `json:"runtime_plan_digest,omitempty"`
			Continuation        valueexecution.Continuation `json:"continuation"`
		}{Schema: valueexecution.ContinuationSchema, Decision: decision, SourcePath: options.filename,
			SourceDigest: plan.SourceDigest, SemanticFingerprint: plan.SemanticFingerprint,
			RuntimePlanDigest: runtimePlanDigest, Continuation: trace}
		if encodeErr := json.NewEncoder(stdout).Encode(payload); encodeErr != nil {
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintf(stderr, "%s: continuation: %v\n", options.filename, err)
	} else {
		fmt.Fprintf(stdout, "executed continuation: iterations=%d/%d feedback_deliveries=%d digest=%s\n",
			trace.IterationsCompleted, trace.IterationsRequested, trace.FeedbackDeliveries, trace.Digest)
	}
	return code
}
