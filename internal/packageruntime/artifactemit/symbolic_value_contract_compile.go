package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func CompileSymbolicValueContract(artifactJSON []byte, subjectSHA string) (SymbolicValueContract, error) {
	var input symbolicValueArtifactInput
	if err := json.Unmarshal(artifactJSON, &input); err != nil {
		return SymbolicValueContract{}, fmt.Errorf("decode symbolic artifact: %w", err)
	}
	if err := validateSymbolicValueArtifact(input, subjectSHA); err != nil {
		return SymbolicValueContract{}, err
	}
	acceptVector, rejectVector, err := bindSymbolicValueVectors(&input)
	if err != nil {
		return SymbolicValueContract{}, err
	}
	contract := buildSymbolicValueContract(input, subjectSHA, acceptVector, rejectVector)
	canonical, err := canonicalSymbolicValueContract(contract)
	if err != nil {
		return SymbolicValueContract{}, fmt.Errorf("canonicalize symbolic value contract: %w", err)
	}
	digest := sha256.Sum256(canonical)
	contract.Digest = "sha256:" + hex.EncodeToString(digest[:])
	return contract, nil
}
