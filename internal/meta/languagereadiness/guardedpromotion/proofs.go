package guardedpromotion

import (
	"crypto/sha256"
	"encoding/hex"
)

func Proofs(coordinates []Coordinate) []Proof {
	return []Proof{
		proof("FOUNDATION", "bind-predecessor-promotion-foundation", coordinates[:5]),
		proof("COHERENCE", "cohere-ci-subject-and-receipt", coordinates[5:8]),
		proof("REGRESSION", "reject-unmerged-or-writing-promotion", coordinates[8:]),
	}
}

func proof(choice, operation string, coordinates []Coordinate) Proof {
	satisfied := true
	hash := sha256.New()
	for _, coordinate := range coordinates {
		hash.Write([]byte(coordinate.ID))
		hash.Write([]byte{0})
		hash.Write([]byte(coordinate.Status))
		if coordinate.Status != statusSatisfied {
			satisfied = false
		}
	}
	return Proof{
		Choice: choice, MetaOperation: operation, Satisfied: satisfied,
		EvidenceDigest: "sha256:" + hex.EncodeToString(hash.Sum(nil)),
	}
}
