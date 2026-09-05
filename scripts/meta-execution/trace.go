package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

const metaExecutionTraceSchema = "gooo/meta-execution-driver-boundary/v1"

type metaExecutionTrace struct {
	headSHA       string
	planDigest    string
	manifestDigest string
	action        generation.Action
	sequence      int
}

type metaExecutionTraceEvent struct {
	Schema                       string `json:"schema"`
	HeadSHA                      string `json:"head_sha"`
	PlanDigest                   string `json:"plan_digest"`
	ManifestDigest               string `json:"manifest_digest"`
	ActionIndicatorID            string `json:"action_indicator_id"`
	Activity                     string `json:"activity"`
	MetaOperation                string `json:"meta_operation"`
	Subject                      string `json:"subject"`
	Output                       string `json:"output"`
	Executor                     string `json:"executor"`
	Evaluator                    string `json:"evaluator"`
	InputContractSourceDigest    string `json:"input_contract_source_digest"`
	InputContractSemanticDigest  string `json:"input_contract_semantic_digest"`
	ContractDigest               string `json:"contract_digest"`
	OperationID                  string `json:"operation_id"`
	OperationSequence            int    `json:"operation_sequence"`
	Pass                         string `json:"pass"`
	Boundary                     string `json:"boundary"`
	CommandKind                  string `json:"command_kind"`
	ObservationScope             string `json:"observation_scope"`
	DiagnosticOnly               bool   `json:"diagnostic_only"`
	SemanticEffect               string `json:"semantic_effect"`
	Permission                   string `json:"permission"`
	ExitCode                     *int   `json:"exit_code,omitempty"`
	ReturnErrorObserved          bool   `json:"return_error_observed,omitempty"`
}

func newMetaExecutionTrace(plan generation.Plan, manifest generation.ExecutionManifest, action generation.Action, sequence int) metaExecutionTrace {
	return metaExecutionTrace{
		headSHA:        plan.HeadSHA,
		planDigest:     plan.PlanDigest,
		manifestDigest: manifest.ManifestDigest,
		action:         action,
		sequence:       sequence,
	}
}

func (trace metaExecutionTrace) emitActionEntered() {
	trace.emit("ACTION_ENTERED", "selected", "action", "UNOBSERVED", "UNOBSERVED", nil, false)
}

func (trace metaExecutionTrace) emitActionReturned(materialized operationMaterialization, runErr *operationError) {
	contractDigest := materialized.ContractDigest
	operationID := materialized.OperationID
	if contractDigest == "" {
		contractDigest = "UNOBSERVED"
	}
	if operationID == "" {
		operationID = "UNOBSERVED"
	}
	trace.emit("ACTION_RETURNED", "selected", "action", contractDigest, operationID, nil, runErr != nil)
}

func (trace metaExecutionTrace) emitProcessCallEntered(pass, commandKind string) {
	trace.emit("PROCESS_CALL_ENTERED", pass, commandKind, "UNOBSERVED", "UNOBSERVED", nil, false)
}

func (trace metaExecutionTrace) emitProcessReturned(pass, commandKind string, observation generation.ProcessObservation, runErr error) {
	exitCode := observation.ExitCode
	trace.emit("PROCESS_RETURNED", pass, commandKind, "UNOBSERVED", "UNOBSERVED", &exitCode, runErr != nil)
}

func (trace metaExecutionTrace) emit(boundary, pass, commandKind, contractDigest, operationID string, exitCode *int, returnErrorObserved bool) {
	event := trace.event(boundary, pass, commandKind, contractDigest, operationID, exitCode, returnErrorObserved)
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, string(payload))
}

func (trace metaExecutionTrace) event(boundary, pass, commandKind, contractDigest, operationID string, exitCode *int, returnErrorObserved bool) metaExecutionTraceEvent {
	if contractDigest == "" {
		contractDigest = "UNOBSERVED"
	}
	if operationID == "" {
		operationID = "UNOBSERVED"
	}
	return metaExecutionTraceEvent{
		Schema:                      metaExecutionTraceSchema,
		HeadSHA:                     trace.headSHA,
		PlanDigest:                  trace.planDigest,
		ManifestDigest:              trace.manifestDigest,
		ActionIndicatorID:           trace.action.IndicatorID,
		Activity:                    trace.action.Activity,
		MetaOperation:               string(trace.action.Operation),
		Subject:                     trace.action.Subject,
		Output:                      trace.action.Output,
		Executor:                    trace.action.Executor,
		Evaluator:                   trace.action.Evaluator,
		InputContractSourceDigest:   trace.action.InputContractSourceDigest,
		InputContractSemanticDigest: trace.action.InputContractSemanticDigest,
		ContractDigest:              contractDigest,
		OperationID:                 operationID,
		OperationSequence:           trace.sequence,
		Pass:                        pass,
		Boundary:                    boundary,
		CommandKind:                 commandKind,
		ObservationScope:            "PROCESS_AND_ACTION_LIFECYCLE_ONLY",
		DiagnosticOnly:              true,
		SemanticEffect:              "UNOBSERVED",
		Permission:                  "UNOBSERVED",
		ExitCode:                    exitCode,
		ReturnErrorObserved:         returnErrorObserved,
	}
}
