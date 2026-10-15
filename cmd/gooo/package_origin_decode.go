package main

import (
	"encoding/json"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bodyexecution"
)

// Choose the latest saved runtime, never construction attempts or history rows.
// This decoder only reads records; it does not replay their producer programs.
func decodePackageOriginRuntime(raw []byte) (bodyexecution.CompositionRuntime, error) {
	var envelope map[string]json.RawMessage
	if err := bodyexecution.DecodeExecutionReceipt(raw, &envelope); err != nil {
		return bodyexecution.CompositionRuntime{}, err
	}
	modes := 0
	for _, key := range []string{"result", "evaluation", "runtime"} {
		if envelope[key] != nil {
			modes++
		}
	}
	if modes > 1 {
		return bodyexecution.CompositionRuntime{}, fmt.Errorf("origin receipt has ambiguous runtime envelopes")
	}
	switch {
	case envelope["result"] != nil:
		return decodePackageWorkspaceOrigins(raw)
	case envelope["evaluation"] != nil:
		var saved bodyConstructOutput
		err := bodyexecution.DecodeExecutionReceipt(raw, &saved)
		return saved.Evaluation.Runtime, err
	case envelope["runtime"] != nil && envelope["composition"] != nil:
		var saved bodyCompositionOutput
		err := bodyexecution.DecodeExecutionReceipt(raw, &saved)
		return saved.Runtime, err
	case envelope["runtime"] != nil:
		var saved bodyexecution.JointEvaluation
		err := bodyexecution.DecodeExecutionReceipt(raw, &saved)
		return saved.Runtime, err
	default:
		var saved bodyexecution.CompositionRuntime
		err := bodyexecution.DecodeExecutionReceipt(raw, &saved)
		return saved, err
	}
}

func decodePackageWorkspaceOrigins(raw []byte) (bodyexecution.CompositionRuntime, error) {
	var saved packageExecutionReceipt
	if err := bodyexecution.DecodeExecutionReceipt(raw, &saved); err != nil {
		return bodyexecution.CompositionRuntime{}, err
	}
	if saved.Schema != "gooo/workspace-body-execution-receipt/v1" || saved.Result == nil || saved.Error != "" {
		return bodyexecution.CompositionRuntime{}, fmt.Errorf("origin input requires a saved workspace execution with its runtime")
	}
	return saved.Result.Runtime, nil
}
