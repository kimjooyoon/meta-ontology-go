package proofchoicealgebra

import (
	"encoding/json"
	"fmt"
)

type Artifact struct {
	Schema         string `json:"schema"`
	SemanticDigest string `json:"semantic_digest"`
	Canonical      string `json:"canonical"`
}

func makeArtifact(lowered lowered) Artifact {
	return Artifact{Schema: "gooo/proof-choice-canonical-artifact/v1", SemanticDigest: lowered.SemanticDigest, Canonical: lowered.Canonical}
}

func decodeArtifact(data []byte) (Artifact, error) {
	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil || artifact.Schema != "gooo/proof-choice-canonical-artifact/v1" || artifact.SemanticDigest == "" || artifact.Canonical == "" {
		return Artifact{}, fmt.Errorf("ARTIFACT_UNKNOWN")
	}
	if digestBytes([]byte(artifact.Canonical)) != artifact.SemanticDigest {
		return Artifact{}, fmt.Errorf("ARTIFACT_DIGEST_MISMATCH")
	}
	return artifact, nil
}

func CanonicalArtifact(path string, source []byte) ([]byte, error) {
	lowered, err := lowerSource(path, source)
	if err != nil {
		return nil, err
	}
	return json.Marshal(makeArtifact(lowered))
}
