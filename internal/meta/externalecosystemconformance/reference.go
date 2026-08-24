package externalecosystemconformance

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed evidence/gomacro.json
var referenceJSON []byte

func Reference() (Capsule, error) {
	var capsule Capsule
	if err := json.Unmarshal(referenceJSON, &capsule); err != nil {
		return Capsule{}, fmt.Errorf("decode reference capsule: %w", err)
	}
	return capsule, nil
}

func cloneCapsule(source Capsule) Capsule {
	clone := source
	clone.Documents = append([]Document(nil), source.Documents...)
	clone.Capabilities = append([]Capability(nil), source.Capabilities...)
	return clone
}

func cloneEvidence(source Evidence) Evidence {
	clone := source
	clone.Readme = append([]byte(nil), source.Readme...)
	clone.GoMod = append([]byte(nil), source.GoMod...)
	return clone
}
