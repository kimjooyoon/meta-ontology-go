package main

import (
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func validatePreparedSplitSource(workspace string, plan generation.Plan, action generation.Action) *operationError {
	if len(plan.GoSplitPreparations) == 0 && plan.GoSplitPolicySHA256 == "" {
		return nil
	}
	raw, err := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(action.Subject)))
	if err == nil {
		err = generation.ValidatePreparedGoSplitSource(plan, action, raw)
	}
	if err == nil {
		return nil
	}
	failure := newOperationError("prepare-workspace", "bind-go-split-source", "GO_SPLIT_PREREQUISITE_STALE",
		"STALE", "observe-source-prerequisites")
	failure.diagnostics = []string{err.Error()}
	return failure
}
