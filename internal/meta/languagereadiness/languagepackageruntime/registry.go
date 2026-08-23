package languagepackageruntime

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Definition struct {
	ID            string `json:"id"`
	Kind          string `json:"kind"`
	Assertion     string `json:"assertion,omitempty"`
	Mutation      string `json:"mutation,omitempty"`
	ExpectedCode  string `json:"expected_code,omitempty"`
	ProofChoice   string `json:"proof_choice"`
	MetaOperation string `json:"meta_operation"`
}

type Registry struct {
	Schema  string       `json:"schema"`
	Version string       `json:"version"`
	Cases   []Definition `json:"cases"`
}

func DecodeRegistry(raw []byte) (Registry, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	registry := Registry{}
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Registry{}, fmt.Errorf("registry has trailing JSON")
	}
	return registry, registry.Validate()
}
