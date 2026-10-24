package main

import (
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcesplit"
)

func checkSplitPrerequisite(workspace, metricsPath string, plan generation.Plan, action generation.Action) *operationError {
	raw, err := os.ReadFile(metricsPath)
	if err != nil {
		return newOperationError("plan-operation", "inspect-source-split", "SOURCE_METRICS_UNAVAILABLE", "DIRECT_MISSING", "restore-source-metrics")
	}
	observe, err := sourcesplit.NewObserver(workspace, plan.HeadSHA, raw)
	if err != nil {
		return newOperationError("plan-operation", "inspect-source-split", "PREREQUISITE_INPUT_UNBOUND", "STALE", "observe-source-prerequisites")
	}
	observed := observe(action)
	for _, previous := range plan.Prerequisites {
		if previous.IndicatorID == action.IndicatorID && previous != observed {
			return newOperationError("plan-operation", "inspect-source-split", "PREREQUISITE_INPUT_CHANGED", "STALE", "observe-source-prerequisites")
		}
	}
	if observed.State == "READY" {
		return nil
	}
	class := "DIRECT_MISSING"
	if observed.State == "BLOCKED" {
		class = "KNOWN_CONTRADICTION"
	} else if observed.State == "DEFERRED" {
		class = "DEPENDENCY_BLOCKED"
	}
	return newOperationError("plan-operation", "inspect-source-split", observed.Reason, class, observed.NextOperation)
}

func splitSourceMatchesPrerequisite(plan generation.Plan, action generation.Action, source []byte) bool {
	for _, observed := range plan.Prerequisites {
		if observed.IndicatorID == action.IndicatorID && observed.SourceDigest != trimDigestPrefix(digestBytes(source)) {
			return false
		}
	}
	return true
}
