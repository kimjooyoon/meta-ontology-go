package pathclosure

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const CodeR4ProofValid = "PROOF_VALID_FINITE_BOUNDARY"

func r4Result(status Status, code, reason string, required []semantic.ID, cost int) R4Result {
	return R4Result{Status: status, Code: code, Reason: reason, RequiredPathIDs: sortedR4IDs(required), Cost: cost}
}
func r4Fail(code, reason string, required []semantic.ID, cost int) R4Result {
	return r4Result(FAIL_CLOSED, code, reason, required, cost)
}
func r4Unknown(code, reason string, required []semantic.ID, cost int) R4Result {
	return r4Result(UNKNOWN, code, reason, required, cost)
}
func invalidR4ID(value semantic.ID, label string) error {
	if _, err := semantic.ParseIdentity(value.String()); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}
func duplicateR4IDs(values []semantic.ID) semantic.ID {
	seen := map[semantic.ID]struct{}{}
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return value
		}
		seen[value] = struct{}{}
	}
	return ""
}
