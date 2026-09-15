package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func validateSymbolicValueReachabilityInputs(artifact Artifact, contract SymbolicValueContract, subjectSHA string) error {
	if !validSymbolicValueHexDigest(subjectSHA, 20) {
		return fmt.Errorf("subject sha must be 40 lowercase hexadecimal characters")
	}
	if artifact.Schema != "gooo/symbolic-invocation-schema-artifact/v1" || artifact.Decision != "PASS" ||
		artifact.Kind != "symbolic-invocation-schema" || artifact.JSONSchema == nil {
		return fmt.Errorf("symbolic artifact identity is not accepted")
	}
	if !validSymbolicValueSHA256(artifact.Digest) || artifact.Effects.RepositoryWrites != 0 || artifact.Effects.MutationAuthority {
		return fmt.Errorf("symbolic artifact must be digest-bound and read-only")
	}
	if contract.Schema != symbolicValueContractSchema || contract.SubjectSHA != subjectSHA ||
		contract.Decision != "PASS" || contract.Resolution != "VALUE_CONTRACT_ONLY" {
		return fmt.Errorf("symbolic value contract identity is not accepted")
	}
	if contract.SourceArtifactDigest != artifact.Digest || !validSymbolicValueSHA256(contract.Digest) {
		return fmt.Errorf("symbolic value contract source binding is not accepted")
	}
	if contract.Effects.RepositoryWrites != 0 || contract.Effects.MutationAuthority || contract.PromotionCreditBPS != 0 {
		return fmt.Errorf("symbolic value contract must remain read-only and non-promoting")
	}
	canonical, err := canonicalSymbolicValueContract(contract)
	if err != nil {
		return fmt.Errorf("canonicalize symbolic value contract input: %w", err)
	}
	digest := sha256.Sum256(canonical)
	if contract.Digest != "sha256:"+hex.EncodeToString(digest[:]) {
		return fmt.Errorf("symbolic value contract digest does not match its payload")
	}
	return nil
}
