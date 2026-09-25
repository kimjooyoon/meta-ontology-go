package selfimprovementloop

import (
	"fmt"
	"strings"
)

func validateInput(in Input) error {
	if in.Schema != Schema {
		return fmt.Errorf("input schema = %q, want %q", in.Schema, Schema)
	}
	if strings.TrimSpace(in.Scenario) == "" || strings.TrimSpace(in.SourceDigest) == "" || strings.TrimSpace(in.ToolchainDigest) == "" {
		return fmt.Errorf("scenario, source digest, and toolchain digest are required")
	}
	return nil
}

func sameMetric(in Input) bool {
	baseline, target := strings.TrimSpace(in.Baseline.Metric), strings.TrimSpace(in.Target.Metric)
	return baseline != "" && baseline == target
}

func unknown(stage, step, reason, class, next, blocked string) *UnknownState {
	return &UnknownState{
		Stage: stage, Step: step, Reason: reason, UnknownClass: class,
		NextOperation: next, BlockedBy: blocked,
	}
}

func nonEmpty(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}
