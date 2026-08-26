package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func CompileSymbolicValueReachability(artifactJSON, contractJSON []byte, subjectSHA string) (SymbolicValueReachability, error) {
	var artifact Artifact
	if err := json.Unmarshal(artifactJSON, &artifact); err != nil {
		return SymbolicValueReachability{}, fmt.Errorf("decode symbolic artifact: %w", err)
	}
	var contract SymbolicValueContract
	if err := json.Unmarshal(contractJSON, &contract); err != nil {
		return SymbolicValueReachability{}, fmt.Errorf("decode symbolic value contract: %w", err)
	}
	if err := validateSymbolicValueReachabilityInputs(artifact, contract, subjectSHA); err != nil {
		return SymbolicValueReachability{}, err
	}
	analysis := analyzeSymbolicValueReachability(artifact, contract)
	reachability := buildSymbolicValueReachability(artifact, contract, subjectSHA, analysis)
	canonical, err := canonicalSymbolicValueReachability(reachability)
	if err != nil {
		return SymbolicValueReachability{}, fmt.Errorf("canonicalize symbolic value reachability: %w", err)
	}
	digest := sha256.Sum256(canonical)
	reachability.Digest = "sha256:" + hex.EncodeToString(digest[:])
	return reachability, nil
}
