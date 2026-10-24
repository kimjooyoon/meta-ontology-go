package main

import (
	"os"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcesplit"
)

func checkSourcePrerequisite(workspace, metricsPath string, plan generation.Plan, action generation.Action) *operationError {
	step := "inspect-source-split"
	if action.Operation == sourcepolicy.OperationExtractFunction {
		step = "inspect-function-extraction"
	}
	raw, err := os.ReadFile(metricsPath)
	if err != nil {
		return newOperationError("plan-operation", step, "SOURCE_METRICS_UNAVAILABLE", "DIRECT_MISSING", "restore-source-metrics")
	}
	observe, err := sourcesplit.NewObserver(workspace, plan.HeadSHA, raw)
	if err != nil {
		return newOperationError("plan-operation", step, "PREREQUISITE_INPUT_UNBOUND", "STALE", "observe-source-prerequisites")
	}
	observed := observe(action)
	for _, previous := range plan.Prerequisites {
		if previous.IndicatorID == action.IndicatorID && !reflect.DeepEqual(previous, observed) {
			return newOperationError("plan-operation", step, "PREREQUISITE_INPUT_CHANGED", "STALE", "observe-source-prerequisites")
		}
	}
	if observed.State == "READY" {
		return nil
	}
	return sourcePrerequisiteFailure(observed, step)
}

func sourcePrerequisiteFailure(observed generation.CandidatePrerequisite, step string) *operationError {
	class := "DIRECT_MISSING"
	if observed.State == "BLOCKED" {
		class = "KNOWN_CONTRADICTION"
	} else if observed.State == "DEFERRED" {
		class = "DEPENDENCY_BLOCKED"
	}
	if observed.UnknownClass != "" {
		class = observed.UnknownClass
	}
	stage := "plan-operation"
	if observed.Stage != "" {
		stage, step = observed.Stage, observed.Step
	}
	failure := newOperationError(stage, step, observed.Reason, class, observed.NextOperation)
	failure.blockedBy = append([]string{}, observed.BlockedBy...)
	failure.diagnostics = append([]string{}, observed.Diagnostics...)
	return failure
}

func sourceMatchesPrerequisite(plan generation.Plan, action generation.Action, source []byte) bool {
	for _, observed := range plan.Prerequisites {
		if observed.IndicatorID == action.IndicatorID && observed.SourceDigest != trimDigestPrefix(digestBytes(source)) {
			return false
		}
	}
	return true
}
