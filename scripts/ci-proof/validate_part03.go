package main

import (
	"fmt"
)

func validateProofDigests(digests proofDigests) error {
	for _, digest := range []string{digests.Source, digests.Semantic, digests.Provenance, digests.Projection, digests.Build, digests.Policy, digests.Schema, digests.Toolchain, digests.Target, digests.Bundle} {
		if !validDigest(digest) {
			return fmt.Errorf("proof digest is missing, zero, or malformed")
		}
	}
	return nil
}
