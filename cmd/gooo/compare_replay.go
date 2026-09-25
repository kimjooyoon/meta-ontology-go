package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

const compareReplayUsage = "usage: gooo compare [--json] <baseline-receipt.json> <candidate-receipt.json>"

type replayReceipt struct {
	Decision  string                   `json:"decision"`
	Execution valueexecution.Execution `json:"execution"`
}

func runCompareReplay(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	if len(args) != 2 {
		fmt.Fprintln(stderr, compareReplayUsage)
		return exitUsage
	}
	baseline, err := readReplayReceipt(args[0], reader)
	if err != nil {
		return reportCompareReplayFailure(jsonMode, stdout, stderr, err)
	}
	candidate, err := readReplayReceipt(args[1], reader)
	if err != nil {
		return reportCompareReplayFailure(jsonMode, stdout, stderr, err)
	}
	comparison := valueexecution.CompareReplay(baseline.Execution, candidate.Execution)
	if jsonMode {
		if err := json.NewEncoder(stdout).Encode(comparison); err != nil {
			return exitFailure
		}
	} else {
		fmt.Fprintf(stdout, "replay: state=%s reason=%s same_plan=%t same_input=%t same_execution=%t\n", comparison.State, comparison.Reason, comparison.SamePlan, comparison.SameInput, comparison.SameExecution)
	}
	if comparison.State != valueexecution.ReplayClosed {
		return exitFailure
	}
	return exitOK
}

func readReplayReceipt(filename string, reader SourceReader) (replayReceipt, error) {
	data, err := reader.ReadFile(filename)
	if err != nil {
		return replayReceipt{}, fmt.Errorf("compare: read %s: %w", filename, err)
	}
	var receipt replayReceipt
	if err := json.Unmarshal(data, &receipt); err != nil {
		return replayReceipt{}, fmt.Errorf("compare: decode %s: %w", filename, err)
	}
	if receipt.Decision != "PASS" {
		return replayReceipt{}, fmt.Errorf("compare: %s is not a PASS receipt", filename)
	}
	return receipt, nil
}

func reportCompareReplayFailure(jsonMode bool, stdout, stderr io.Writer, err error) int {
	if jsonMode {
		_ = json.NewEncoder(stdout).Encode(map[string]string{
			"schema": compareReplayUsage,
			"state":  string(valueexecution.ReplayUnknown),
			"reason": err.Error(),
		})
	} else {
		fmt.Fprintln(stderr, err)
	}
	return exitFailure
}
