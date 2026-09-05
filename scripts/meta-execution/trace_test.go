package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestMetaExecutionTraceBindsActionAndKeepsLifecycleNonSemantic(t *testing.T) {
	action := generation.Action{
		IndicatorID:                 "indicator-1",
		Activity:                    "activity-1",
		Operation:                   sourcepolicy.OperationExtractFunction,
		Subject:                     "subject-1",
		Output:                      "output-1",
		Executor:                    "executor-1",
		Evaluator:                   "evaluator-1",
		InputContractSourceDigest:   "source-contract-1",
		InputContractSemanticDigest: "semantic-contract-1",
	}
	trace := newMetaExecutionTrace(
		generation.Plan{HeadSHA: "head-1", PlanDigest: "plan-1"},
		generation.ExecutionManifest{ManifestDigest: "manifest-1"},
		action,
		2,
		newMetaExecutionTraceStateWithWriter(&bytes.Buffer{}),
	)

	entered := trace.event("PROCESS_CALL_ENTERED", "first", "executor", "", "", "UNOBSERVED", nil, nil)
	if entered.Schema != metaExecutionTraceSchema || entered.HeadSHA != "head-1" || entered.PlanDigest != "plan-1" || entered.ManifestDigest != "manifest-1" {
		t.Fatalf("trace context = %#v", entered)
	}
	if entered.ActionIndicatorID != action.IndicatorID || entered.Activity != action.Activity || entered.MetaOperation != string(action.Operation) || entered.Subject != action.Subject || entered.Output != action.Output || entered.Executor != action.Executor || entered.Evaluator != action.Evaluator {
		t.Fatalf("trace action binding = %#v", entered)
	}
	if entered.InputContractSourceDigest != action.InputContractSourceDigest || entered.InputContractSemanticDigest != action.InputContractSemanticDigest || entered.OperationSequence != 2 || entered.Pass != "first" || entered.Boundary != "PROCESS_CALL_ENTERED" || entered.CommandKind != "executor" {
		t.Fatalf("trace execution binding = %#v", entered)
	}
	if entered.ContractDigest != "UNOBSERVED" || entered.ExitCode != nil || entered.SemanticEffect != "UNOBSERVED" || entered.Permission != "UNOBSERVED" {
		t.Fatalf("trace entry scope = %#v", entered)
	}

	exitCode := 0
	returnErrorObserved := false
	returned := trace.event("PROCESS_RETURNED", "first", "executor", "sha256:contract", "operation-1", "generation.ProcessObservation.ExitCode", &exitCode, &returnErrorObserved)
	if returned.ContractDigest != "sha256:contract" || returned.OperationID != "operation-1" || returned.ExitCode == nil || *returned.ExitCode != 0 || returned.ReturnErrorObserved == nil || *returned.ReturnErrorObserved {
		t.Fatalf("trace return observation = %#v", returned)
	}
	if returned.ExitCodeSource != "generation.ProcessObservation.ExitCode" {
		t.Fatalf("trace exit-code source = %#v", returned)
	}
	if returned.SemanticEffect != "UNOBSERVED" || returned.Permission != "UNOBSERVED" {
		t.Fatalf("trace return semantic scope = %#v", returned)
	}
}

func TestMetaExecutionTraceEmitsInvocationAndEventSequence(t *testing.T) {
	var output bytes.Buffer
	state := newMetaExecutionTraceStateWithWriter(&output)
	action := generation.Action{
		IndicatorID: "indicator-1",
		Activity:    "activity-1",
		Operation:   sourcepolicy.OperationExtractFunction,
		Subject:     "subject-1",
	}
	trace := newMetaExecutionTrace(
		generation.Plan{HeadSHA: "head-1", PlanDigest: "plan-1"},
		generation.ExecutionManifest{ManifestDigest: "manifest-1"},
		action,
		1,
		state,
	)
	secondTrace := newMetaExecutionTrace(
		generation.Plan{HeadSHA: "head-1", PlanDigest: "plan-1"},
		generation.ExecutionManifest{ManifestDigest: "manifest-1"},
		action,
		2,
		state,
	)

	trace.emitActionEntered()
	secondTrace.emitActionEntered()
	events := decodeTraceEvents(t, output.Bytes())
	if len(events) != 2 {
		t.Fatalf("trace events = %#v", events)
	}
	if events[0].InvocationID == "" || events[0].InvocationID != events[1].InvocationID {
		t.Fatalf("trace invocation binding = %#v", events)
	}
	if events[0].EventSequence != 1 || events[1].EventSequence != 2 {
		t.Fatalf("trace event sequence = %#v", events)
	}
}

func TestMetaExecutionTraceDoesNotSynthesizeIncompleteActionSequence(t *testing.T) {
	var output bytes.Buffer
	state := newMetaExecutionTraceStateWithWriter(&output)
	trace := newMetaExecutionTrace(
		generation.Plan{HeadSHA: "head-1", PlanDigest: "plan-1"},
		generation.ExecutionManifest{ManifestDigest: "manifest-1"},
		generation.Action{IndicatorID: "indicator-1", Operation: sourcepolicy.OperationExtractFunction},
		1,
		state,
	)

	trace.emitActionEntered()
	events := decodeTraceEvents(t, output.Bytes())
	if len(events) != 1 || events[0].Boundary != "ACTION_ENTERED" || events[0].EventSequence != 1 {
		t.Fatalf("incomplete trace sequence = %#v", events)
	}
}

func TestObserveProcessCallIgnoresTraceWriterFailure(t *testing.T) {
	trace := newMetaExecutionTrace(
		generation.Plan{HeadSHA: "head-1", PlanDigest: "plan-1"},
		generation.ExecutionManifest{ManifestDigest: "manifest-1"},
		generation.Action{IndicatorID: "indicator-1", Operation: sourcepolicy.OperationExtractFunction},
		1,
		newMetaExecutionTraceStateWithWriter(errorWriter{}),
	)
	wantResult := processResult{Observation: generation.ProcessObservation{ExitCode: 7}, Stdout: []byte("stdout"), Stderr: []byte("stderr")}
	wantErr := errors.New("original process error")

	gotResult, gotErr := observeProcessCall(&trace, "first", "executor", func() (processResult, error) {
		return wantResult, wantErr
	})
	if !reflect.DeepEqual(gotResult, wantResult) {
		t.Fatalf("process result changed by tracing: got %#v want %#v", gotResult, wantResult)
	}
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("process error changed by tracing: got %v want %v", gotErr, wantErr)
	}
}

type errorWriter struct{}

func (errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("trace writer failed")
}

func decodeTraceEvents(t *testing.T, payload []byte) []metaExecutionTraceEvent {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(payload), []byte("\n"))
	if len(lines) == 1 && len(lines[0]) == 0 {
		return nil
	}
	events := make([]metaExecutionTraceEvent, 0, len(lines))
	for _, line := range lines {
		var event metaExecutionTraceEvent
		if err := json.Unmarshal(line, &event); err != nil {
			t.Fatalf("decode trace event %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}
