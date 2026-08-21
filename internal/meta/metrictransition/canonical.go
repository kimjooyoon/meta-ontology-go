package metrictransition

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func sealState(state RepositoryState) (RepositoryState, error) {
	state.Digest = ""
	digest, err := digestValue(state)
	if err != nil {
		return RepositoryState{}, err
	}
	state.Digest = digest
	return state, nil
}

func sealLedger(ledger TransitionLedger) (TransitionLedger, error) {
	ledger.Digest = ""
	digest, err := digestValue(ledger)
	if err != nil {
		return TransitionLedger{}, err
	}
	ledger.Digest = digest
	return ledger, nil
}

func buildEffectEvidence(inputs inputSet) (EffectEvidence, error) {
	artifacts := []ArtifactEvidence{
		{Role: "effect-ledger", Digest: digestBytes(inputs.effect)},
		{Role: "executed-receipts", Digest: digestBytes(inputs.receipts)},
		{Role: "executed-provenance", Digest: digestBytes(inputs.provenance)},
		{Role: "content-patch", Digest: digestBytes(inputs.patch)},
	}
	setDigest, err := digestValue(artifacts)
	if err != nil {
		return EffectEvidence{}, err
	}
	return EffectEvidence{Verifier: "transformationeffect.VerifyFiles", Artifacts: artifacts, SetDigest: setDigest}, nil
}
