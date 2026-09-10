package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

// The trace is diagnostic test output, never canonical operation evidence.
func executeCollapseWithTestTrace(t *testing.T, workspace, gitDir, metricsPath string, plan generation.Plan, action generation.Action) (operationMaterialization, *operationError) {
	t.Helper()
	manifest := generation.BuildExecutionManifest(plan)
	sequence := 0
	for index, step := range manifest.Steps {
		if step.ActionIndicatorID == action.IndicatorID {
			sequence = index + 1
			break
		}
	}
	if sequence == 0 {
		t.Fatal("native collapse trace action is absent from the derived manifest")
	}
	var output bytes.Buffer
	defer func() {
		for event := range strings.SplitSeq(strings.TrimSpace(output.String()), "\n") {
			if event != "" {
				t.Logf("native-collapse-test-trace=%s", event)
			}
		}
	}()
	trace := newMetaExecutionTrace(plan, manifest, action, sequence, newMetaExecutionTraceStateWithWriter(&output))
	trace.emitActionEntered()
	materialized, failure := executeCollapse(workspace, gitDir, metricsPath, plan, action, trace)
	trace.emitActionReturned(materialized, failure)
	return materialized, failure
}
