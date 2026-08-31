package artifactemit

import "fmt"

func bindSymbolicValueVectors(input *symbolicValueArtifactInput) (*symbolicValueVectorInput, *symbolicValueVectorInput, error) {
	var acceptVector *symbolicValueVectorInput
	var rejectVector *symbolicValueVectorInput
	seen := make(map[string]struct{}, len(input.Conformance.Vectors))
	for i := range input.Conformance.Vectors {
		vector := &input.Conformance.Vectors[i]
		if _, duplicate := seen[vector.ID]; duplicate {
			return nil, nil, fmt.Errorf("duplicate symbolic vector %q", vector.ID)
		}
		seen[vector.ID] = struct{}{}
		switch vector.Expected {
		case "ACCEPT", "REJECT":
		default:
			return nil, nil, fmt.Errorf("symbolic vector %q has unknown expected verdict %q", vector.ID, vector.Expected)
		}
		switch vector.ID {
		case "accept-exact":
			acceptVector = vector
		case "reject-missing-activity":
			rejectVector = vector
		default:
			return nil, nil, fmt.Errorf("symbolic vector %q is not contract-bound", vector.ID)
		}
	}
	if acceptVector == nil || rejectVector == nil {
		return nil, nil, fmt.Errorf("required symbolic vectors are missing")
	}
	if err := validateSymbolicValueVectorPair(acceptVector, rejectVector); err != nil {
		return nil, nil, err
	}
	return acceptVector, rejectVector, nil
}
