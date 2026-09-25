package languagediagnosticprovenancebinding

import "fmt"

func obligationStatus(input Input) string {
	for _, obligation := range input.Readiness.Obligations {
		if obligation.ID == "LANGUAGE-DIAGNOSTIC-PROVENANCE" {
			return obligation.Status
		}
	}
	return "UNKNOWN"
}

func proofLabel(input Input) string {
	passed := 0
	for _, proof := range input.Provenance.Proofs {
		if proof.Passed {
			passed++
		}
	}
	return fmt.Sprintf("%d/%d", passed, len(input.Provenance.Proofs))
}

func proofsBound(input Input) bool {
	if len(input.Provenance.Proofs) != 3 {
		return false
	}
	choices := map[string]bool{}
	for _, proof := range input.Provenance.Proofs {
		choices[proof.Choice] = proof.Passed
	}
	return choices["FOUNDATION"] && choices["COHERENCE"] && choices["REGRESSION"]
}

func sealedLabel(input Input) string {
	return fmt.Sprintf("%d/%d/%d/%d", input.Provenance.Summary.Unresolved,
		input.Provenance.Summary.EffectfulStages, input.Provenance.RepositoryWrites,
		boolCount(input.Provenance.MutationAuthorized))
}

func sealedBound(input Input) bool {
	return input.Provenance.Summary.Unresolved == 0 &&
		input.Provenance.Summary.EffectfulStages == 0 &&
		input.Provenance.RepositoryWrites == 0 && !input.Provenance.MutationAuthorized
}
