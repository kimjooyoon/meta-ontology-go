package causalci

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return digestBytes(data), nil
}

func transitionDigest(value ClaimTransition) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func receiptDigest(value Receipt) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}

func planDigest(value Receipt) (string, error) {
	projection := struct {
		ObservationDigest string              `json:"observation_digest"`
		Operation         Operation           `json:"operation"`
		ExecutionMode     string              `json:"execution_mode"`
		Conformance       Conformance         `json:"conformance"`
		Subjects          []SubjectResolution `json:"subjects"`
		ClaimTransitions  []ClaimTransition   `json:"claim_transitions"`
	}{
		ObservationDigest: value.ObservationDigest,
		Operation:         value.Operation,
		ExecutionMode:     value.ExecutionMode,
		Conformance:       value.Conformance,
		Subjects:          value.Subjects,
		ClaimTransitions:  value.ClaimTransitions,
	}
	return digestJSON(projection)
}

func interventionDigest(value InterventionReport) (string, error) {
	value.Digest = ""
	return digestJSON(value)
}
