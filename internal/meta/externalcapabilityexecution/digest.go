package externalcapabilityexecution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func digestBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestValue(value any) string {
	raw, _ := json.Marshal(value)
	return digestBytes(raw)
}

func sealObservation(observation Observation) Observation {
	observation.ObservationDigest = ""
	observation.ObservationDigest = digestValue(observation)
	return observation
}
