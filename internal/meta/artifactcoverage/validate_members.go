package artifactcoverage

import (
	"fmt"
	"strings"
)

func validateIndicators(indicators []Indicator, operations map[string]MetaOperation) error {
	seen, classes := make(map[string]bool), make(map[IndicatorClass]bool)
	for _, item := range indicators {
		operation, exists := operations[item.MetaOperation]
		validClass := item.Class == ClassOutcome || item.Class == ClassDriver || item.Class == ClassGuardrail
		validRelation := item.Relation == RelationGreaterOrEqual || item.Relation == RelationLessOrEqual
		if item.MetricID == "" || item.Unit == "" || item.Producer == "" || item.Consumer == "" ||
			!validClass || !validRelation || !exists || operation.Activity != item.Activity ||
			operation.ProofChoice != item.ProofChoice {
			return fmt.Errorf("incomplete indicator %q", item.MetricID)
		}
		if seen[item.MetricID] {
			return fmt.Errorf("duplicate indicator %q", item.MetricID)
		}
		seen[item.MetricID], classes[item.Class] = true, true
	}
	for _, class := range []IndicatorClass{ClassOutcome, ClassDriver, ClassGuardrail} {
		if !classes[class] {
			return fmt.Errorf("missing %s indicator class", class)
		}
	}
	return nil
}

func validateBindings(bindings []ArtifactBinding) error {
	operations, evidence := make(map[string]bool), make(map[string]bool)
	for _, item := range bindings {
		complete := item.Operation != "" && item.Activity != "" && validProof(item.ProofChoice) &&
			item.Registry != "" && item.Executor != "" && item.Evaluator != "" && item.EvidenceKey != "" &&
			strings.Contains(item.ArtifactPattern, "{head_sha}") && item.ExactHead && item.DigestBound && item.ReplayRequired
		if !complete {
			return fmt.Errorf("incomplete artifact binding %q", item.Operation)
		}
		if operations[item.Operation] || evidence[item.EvidenceKey] {
			return fmt.Errorf("ambiguous artifact binding %q", item.Operation)
		}
		operations[item.Operation], evidence[item.EvidenceKey] = true, true
	}
	return nil
}
