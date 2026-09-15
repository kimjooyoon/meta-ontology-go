package artifactcoverage

import "fmt"

func Validate(program Program) error {
	if program.Schema != Schema || program.AuthorityPath != AuthorityPath {
		return fmt.Errorf("operation artifact program identity is malformed")
	}
	operations, err := validateOperations(program.MetaOperations)
	if err != nil {
		return err
	}
	if err := validateIndicators(program.Indicators, operations); err != nil {
		return err
	}
	return validateBindings(program.ArtifactBindings)
}

func validateOperations(operations []MetaOperation) (map[string]MetaOperation, error) {
	index := make(map[string]MetaOperation, len(operations))
	proofs := make(map[ProofChoice]bool)
	for _, operation := range operations {
		if operation.ID == "" || operation.Activity == "" || !validProof(operation.ProofChoice) {
			return nil, fmt.Errorf("incomplete meta operation %q", operation.ID)
		}
		if _, exists := index[operation.ID]; exists {
			return nil, fmt.Errorf("duplicate meta operation %q", operation.ID)
		}
		index[operation.ID], proofs[operation.ProofChoice] = operation, true
	}
	for _, proof := range []ProofChoice{ProofFoundation, ProofCoherence, ProofRegression} {
		if !proofs[proof] {
			return nil, fmt.Errorf("missing %s proof branch", proof)
		}
	}
	return index, nil
}

func validProof(proof ProofChoice) bool {
	return proof == ProofFoundation || proof == ProofCoherence || proof == ProofRegression
}
