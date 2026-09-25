package guardedcapability

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness/guardedpromotion"

func proofs(coordinates []guardedpromotion.Coordinate) []guardedpromotion.Proof {
	return []guardedpromotion.Proof{
		proof("FOUNDATION", "bind-authorized-promotion-foundation", coordinates),
		proof("COHERENCE", "cohere-foundation-and-current-implementation", coordinates),
		proof("REGRESSION", "reject-drift-or-observer-authority", coordinates),
	}
}

func proof(choice, operation string, coordinates []guardedpromotion.Coordinate) guardedpromotion.Proof {
	selected := make([]guardedpromotion.Coordinate, 0)
	satisfied := true
	for _, item := range coordinates {
		if item.ProofChoice != choice {
			continue
		}
		selected = append(selected, item)
		if item.Status != "SATISFIED" {
			satisfied = false
		}
	}
	return guardedpromotion.Proof{Choice: choice, MetaOperation: operation,
		Satisfied: satisfied, EvidenceDigest: digest(selected)}
}
