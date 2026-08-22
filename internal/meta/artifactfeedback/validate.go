package artifactfeedback

import "fmt"

func Validate(program Program) error {
	if program.Schema != Schema || program.Authority != Authority {
		return fmt.Errorf("artifact feedback program identity is malformed")
	}
	operations := make(map[string]MetaOperation, len(program.MetaOperations))
	proofs := make(map[ProofChoice]bool)
	for _, operation := range program.MetaOperations {
		if operation.ID == "" || operation.Activity == "" || !validProof(operation.ProofChoice) {
			return fmt.Errorf("incomplete meta operation %q", operation.ID)
		}
		if _, exists := operations[operation.ID]; exists {
			return fmt.Errorf("duplicate meta operation %q", operation.ID)
		}
		operations[operation.ID], proofs[operation.ProofChoice] = operation, true
	}
	for _, proof := range []ProofChoice{ProofFoundation, ProofCoherence, ProofRegression} {
		if !proofs[proof] {
			return fmt.Errorf("missing %s proof branch", proof)
		}
	}
	classes, metrics := make(map[IndicatorClass]bool), make(map[string]bool)
	for _, item := range program.Indicators {
		operation, exists := operations[item.MetaOperation]
		validClass := item.Class == ClassOutcome || item.Class == ClassDriver || item.Class == ClassGuardrail
		validRelation := item.Relation == RelationGreaterOrEqual || item.Relation == RelationLessOrEqual
		if item.MetricID == "" || item.Unit == "" || item.Producer == "" || item.Consumer == "" ||
			!validClass || !validRelation || !exists || operation.Activity != item.Activity ||
			operation.ProofChoice != item.ProofChoice {
			return fmt.Errorf("incomplete indicator %q", item.MetricID)
		}
		if metrics[item.MetricID] {
			return fmt.Errorf("duplicate indicator %q", item.MetricID)
		}
		metrics[item.MetricID], classes[item.Class] = true, true
	}
	for _, class := range []IndicatorClass{ClassOutcome, ClassDriver, ClassGuardrail} {
		if !classes[class] {
			return fmt.Errorf("missing %s indicator class", class)
		}
	}
	return nil
}

func validProof(proof ProofChoice) bool {
	return proof == ProofFoundation || proof == ProofCoherence || proof == ProofRegression
}
