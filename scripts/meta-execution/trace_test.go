package main

import (
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
	)

	entered := trace.event("PROCESS_CALL_ENTERED", "first", "executor", "", "", nil, false)
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
	returned := trace.event("PROCESS_RETURNED", "first", "executor", "sha256:contract", "operation-1", &exitCode, false)
	if returned.ContractDigest != "sha256:contract" || returned.OperationID != "operation-1" || returned.ExitCode == nil || *returned.ExitCode != 0 || returned.ReturnErrorObserved {
		t.Fatalf("trace return observation = %#v", returned)
	}
	if returned.SemanticEffect != "UNOBSERVED" || returned.Permission != "UNOBSERVED" {
		t.Fatalf("trace return semantic scope = %#v", returned)
	}
}
