package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type options struct {
	headSHA, branch, conclusion string
	metrics, plan, execution    string
	receipts, provenance        string
	contract, output            string
	runID                       int64
	check                       bool
}

type document[T any] struct {
	Value      T
	FileSHA256 string
}

type inputs struct {
	Metrics    document[metricsDocument]
	Plan       document[planDocument]
	Execution  document[executionDocument]
	Receipts   document[receiptDocument]
	Provenance document[provenanceDocument]
	Contract   document[contractDocument]
}

func loadInputs(opts options) (inputs, error) {
	var in inputs
	var err error
	if in.Metrics, err = decodeDocument[metricsDocument](opts.metrics); err != nil {
		return in, fmt.Errorf("metrics: %w", err)
	}
	if in.Plan, err = decodeDocument[planDocument](opts.plan); err != nil {
		return in, fmt.Errorf("plan: %w", err)
	}
	if in.Execution, err = decodeDocument[executionDocument](opts.execution); err != nil {
		return in, fmt.Errorf("execution: %w", err)
	}
	if in.Receipts, err = decodeDocument[receiptDocument](opts.receipts); err != nil {
		return in, fmt.Errorf("receipts: %w", err)
	}
	if in.Provenance, err = decodeDocument[provenanceDocument](opts.provenance); err != nil {
		return in, fmt.Errorf("provenance: %w", err)
	}
	if in.Contract, err = decodeDocument[contractDocument](opts.contract); err != nil {
		return in, fmt.Errorf("contract: %w", err)
	}
	return in, nil
}

func decodeDocument[T any](path string) (document[T], error) {
	var result document[T]
	data, err := os.ReadFile(path)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(data, &result.Value); err != nil {
		return result, err
	}
	result.FileSHA256 = digestBytes(data)
	return result, nil
}
