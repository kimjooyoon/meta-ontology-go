package artifactemit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func failed(kind, resolution, reason string) Artifact {
	return finish(Artifact{
		Schema: OperationManifestSchema, Decision: "FAIL_CLOSED", Resolution: resolution,
		Reason: reason, Kind: kind, Definitions: DefinitionSet{Language: "gooo", Files: []Definition{}},
		Extensions: registryReceipt(),
	})
}

func finish(artifact Artifact) Artifact {
	artifact.Digest = ""
	payload, _ := json.Marshal(artifact)
	sum := sha256.Sum256(payload)
	artifact.Digest = "sha256:" + hex.EncodeToString(sum[:])
	return artifact
}

func Marshal(artifact Artifact) ([]byte, error) {
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
