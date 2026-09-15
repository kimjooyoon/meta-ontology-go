package semanticresolution

import "fmt"

func Validate(program Program) error {
	if program.Schema != Schema || program.Authority != Authority {
		return fmt.Errorf("semantic resolution identity is malformed")
	}
	if len(program.Resolutions) != 3 || len(program.MetaOperations) != 6 || len(program.Indicators) != 7 {
		return fmt.Errorf("semantic resolution cardinality is malformed")
	}
	if program.Resolutions[0] != ResolutionExactOperation || program.Resolutions[1] != ResolutionOperationClass || program.Resolutions[2] != ResolutionInvariantOnly {
		return fmt.Errorf("semantic resolution order is malformed")
	}
	operations := make(map[string]MetaOperation, len(program.MetaOperations))
	proofs := make(map[ProofChoice]bool, 3)
	for _, operation := range program.MetaOperations {
		if operation.ID == "" || operation.Activity == "" || !validProof(operation.ProofChoice) {
			return fmt.Errorf("semantic resolution operation is malformed")
		}
		if _, exists := operations[operation.ID]; exists {
			return fmt.Errorf("duplicate semantic resolution operation %q", operation.ID)
		}
		operations[operation.ID], proofs[operation.ProofChoice] = operation, true
	}
	seen := make(map[string]bool, len(program.Indicators))
	for _, metric := range program.Indicators {
		operation, exists := operations[metric.MetaOperation]
		if metric.MetricID == "" || seen[metric.MetricID] || !exists || operation.Activity != metric.Activity {
			return fmt.Errorf("semantic resolution indicator binding is malformed")
		}
		if !validClass(metric.Class) || !validRelation(metric.Relation) || !validProof(metric.ProofChoice) || metric.ProofChoice != operation.ProofChoice {
			return fmt.Errorf("semantic resolution indicator semantics are malformed")
		}
		if metric.Unit == "" || metric.Producer == "" || metric.Consumer == "" || metric.Target < 0 {
			return fmt.Errorf("semantic resolution indicator evidence is malformed")
		}
		seen[metric.MetricID] = true
	}
	if !proofs[ProofFoundation] || !proofs[ProofCoherence] || !proofs[ProofRegression] {
		return fmt.Errorf("semantic resolution proof trilemma is incomplete")
	}
	return nil
}

func validProof(value ProofChoice) bool {
	return value == ProofFoundation || value == ProofCoherence || value == ProofRegression
}

func validClass(value IndicatorClass) bool {
	return value == ClassOutcome || value == ClassDriver || value == ClassGuardrail
}

func validRelation(value Relation) bool {
	return value == RelationGreaterOrEqual || value == RelationLessOrEqual
}
