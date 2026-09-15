package artifactemit

import (
	"fmt"
	"slices"
	"strings"
)

func validateSymbolicValueVectorPair(acceptVector, rejectVector *symbolicValueVectorInput) error {
	if acceptVector.Expected != "ACCEPT" || acceptVector.ProofChoice != "FOUNDATION" ||
		acceptVector.MetaOperation != "project-exact-symbolic-invocation" ||
		strings.TrimSpace(acceptVector.Instance.Activity) == "" || len(acceptVector.Instance.Inputs) == 0 {
		return fmt.Errorf("accept vector does not prove a complete symbolic value")
	}
	if rejectVector.Expected != "REJECT" || rejectVector.ProofChoice != "REGRESSION" ||
		rejectVector.MetaOperation != "remove-required-activity" ||
		strings.TrimSpace(rejectVector.Instance.Activity) != "" || len(rejectVector.Instance.Inputs) == 0 {
		return fmt.Errorf("reject vector does not prove the missing-activity boundary")
	}
	if !slices.Equal(acceptVector.Instance.Inputs, rejectVector.Instance.Inputs) {
		return fmt.Errorf("generated vectors do not share an input boundary")
	}
	return nil
}
