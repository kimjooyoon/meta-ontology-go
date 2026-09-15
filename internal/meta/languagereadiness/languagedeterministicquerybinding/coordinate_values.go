package languagedeterministicquerybinding

import "fmt"

func queryObligationStatus(input Input) string {
	for _, obligation := range input.Readiness.Obligations {
		if obligation.ID == "LANGUAGE-DETERMINISTIC-QUERY" {
			return obligation.Status
		}
	}
	return "UNKNOWN"
}

func queryProofLabel(input Input) string {
	passed := 0
	for _, proof := range input.Query.Proofs {
		if proof.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d", passed, len(input.Query.Proofs))
}

func queryProofsBound(input Input) bool {
	if len(input.Query.Proofs) != 3 {
		return false
	}
	choices := map[string]bool{}
	for _, proof := range input.Query.Proofs {
		choices[proof.Choice] = proof.Passed
	}
	return choices["FOUNDATION"] && choices["COHERENCE"] && choices["REGRESSION"]
}

func sealedLabel(input Input) string {
	return fmt.Sprintf("%d/%d/%d/%d", input.Query.Summary.Unresolved,
		input.Query.Summary.EffectfulStages, input.Query.RepositoryWrites, boolCount(input.Query.MutationAuthorized))
}

func sealedBound(input Input) bool {
	return input.Query.Summary.Unresolved == 0 && input.Query.Summary.EffectfulStages == 0 &&
		input.Query.RepositoryWrites == 0 && !input.Query.MutationAuthorized
}
