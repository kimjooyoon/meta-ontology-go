package languagegointeroperationbinding

import "fmt"

func interopObligationStatus(input Input) string {
	for _, obligation := range input.Readiness.Obligations {
		if obligation.ID == "LANGUAGE-GO-INTEROPERATION" {
			return obligation.Status
		}
	}
	return "UNKNOWN"
}

func proofLabel(input Input) string {
	passed := 0
	for _, proof := range input.Interoperation.Proofs {
		if proof.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d", passed, len(input.Interoperation.Proofs))
}

func proofsBound(input Input) bool {
	if len(input.Interoperation.Proofs) != 3 {
		return false
	}
	choices := map[string]bool{}
	for _, proof := range input.Interoperation.Proofs {
		choices[proof.Choice] = proof.Passed
	}
	return choices["FOUNDATION"] && choices["COHERENCE"] && choices["REGRESSION"]
}

func sealedLabel(input Input) string {
	return fmt.Sprintf("%d/%d/%d/%d", input.Interoperation.Summary.Unresolved,
		input.Interoperation.Summary.EffectfulStages, input.Interoperation.RepositoryWrites,
		boolCount(input.Interoperation.MutationAuthorized))
}

func sealedBound(input Input) bool {
	return input.Interoperation.Summary.Unresolved == 0 && input.Interoperation.Summary.EffectfulStages == 0 &&
		input.Interoperation.RepositoryWrites == 0 && !input.Interoperation.MutationAuthorized
}
