package toolchainrelease

import "fmt"

func validateProofChoices(proofs []Proof) error {
	choices := map[string]int{}
	for _, proof := range proofs {
		choices[proof.ProofChoice]++
		if proof.EvidenceDigest == "" {
			return fmt.Errorf("TOOLCHAIN_RELEASE_PROOF_EVIDENCE_MISSING")
		}
	}
	if choices["FOUNDATION"] != 1 || choices["COHERENCE"] != 1 || choices["REGRESSION"] != 1 {
		return fmt.Errorf("TOOLCHAIN_RELEASE_PROOF_PARTITION_MISMATCH")
	}
	return nil
}
