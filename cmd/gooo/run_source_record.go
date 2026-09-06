package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

type recordPlanReport struct {
	Schema              string                         `json:"schema"`
	Decision            string                         `json:"decision"`
	Scope               string                         `json:"scope"`
	SemanticAdmission   string                         `json:"semantic_admission"`
	SourcePath          string                         `json:"source_path"`
	SourceDigest        string                         `json:"source_digest"`
	SemanticFingerprint string                         `json:"semantic_fingerprint"`
	Entry               string                         `json:"entry"`
	Execution           valueexecution.RecordExecution `json:"execution"`
	Failure             *valueexecution.Failure        `json:"failure,omitempty"`
}

func runSourceRecordPlan(options runSourceOptions, source []byte, reader SourceReader, jsonMode bool, stdout, stderr io.Writer) int {
	plan, execution, err := executeSourceRecordPlan(options, source, reader)
	report := recordPlanReport{
		Schema: "gooo/record-execution-plan/v1", Decision: "PASS", Scope: valueexecution.RecordTransportScope,
		SemanticAdmission: "UNASSESSED", SourcePath: options.filename, SourceDigest: plan.SourceDigest,
		SemanticFingerprint: plan.SemanticFingerprint, Entry: options.entry, Execution: execution,
	}
	code := exitOK
	if err != nil {
		code, report.Decision = exitFailure, "FAIL_CLOSED"
		failure, found := valueexecution.FailureOf(err)
		if !found {
			failure = valueexecution.Failure{Code: valueexecution.Reason(err), Stage: "EXECUTE", Step: "record-plan", Detail: err.Error()}
		}
		report.Failure = &failure
	}
	if jsonMode {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			return exitFailure
		}
	} else if err != nil {
		fmt.Fprintf(stderr, "%s: record plan: %v\n", options.filename, err)
	} else {
		fmt.Fprintf(stdout, "executed record plan: entry=%s activities=%d applies=%d deliveries=%d scope=%s semantic_admission=UNASSESSED\n",
			options.entry, len(execution.Activities), execution.ApplyCalls, execution.Deliveries, report.Scope)
	}
	return code
}

func executeSourceRecordPlan(options runSourceOptions, source []byte, reader SourceReader) (valueexecution.RecordPlan, valueexecution.RecordExecution, error) {
	empty := valueexecution.RecordExecution{Scope: valueexecution.RecordTransportScope}
	raw, err := reader.ReadFile(options.input)
	if err != nil {
		return valueexecution.RecordPlan{}, empty, valueexecution.Failure{
			Code: valueexecution.ReasonSourceReadFailed, Stage: "INPUT", Step: "read-record-input", Detail: err.Error(),
		}
	}
	fields, err := valueexecution.DecodeRecordInput(raw)
	if err != nil {
		return valueexecution.RecordPlan{}, empty, err
	}
	plan, err := valueexecution.CompileRecordPlan(options.filename, source)
	if err != nil {
		return plan, empty, err
	}
	execution, err := plan.Execute(map[string]valueexecution.RecordFields{options.entry: fields})
	return plan, execution, err
}
