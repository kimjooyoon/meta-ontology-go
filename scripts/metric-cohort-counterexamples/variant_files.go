package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"

	artifact "github.com/kimjooyoon/meta-ontology-go/internal/meta/metriccounterfactualio"
)

type inputReference struct {
	Member string `json:"member"`
	State  string `json:"state"`
	Digest string `json:"digest"`
}

func saveVariant(directory string, source origin, baseline fixture, v variant) error {
	if err := os.MkdirAll(directory, 0750); err != nil {
		return err
	}
	original, payloads := inputPayloads(baseline.Inputs), inputPayloads(v.Inputs)
	original["concept-operation-binding.json"], payloads["concept-operation-binding.json"] = baseline.Raw, v.Receipt
	names := make([]string, 0, len(payloads))
	for name := range payloads {
		names = append(names, name)
	}
	slices.Sort(names)
	references := make([]inputReference, 0, len(names))
	for _, name := range names {
		raw, state := payloads[name], "CHANGED_MEMBER"
		if raw == nil {
			state = "ABSENT_PAYLOAD"
		} else if bytes.Equal(raw, original[name]) {
			state = "ORIGINAL_PRODUCER_ARCHIVE"
		} else if err := os.WriteFile(filepath.Join(directory, name), raw, 0640); err != nil {
			return err
		}
		references = append(references, inputReference{name, state, artifact.ContentDigest(raw)})
	}
	manifest := struct {
		Schema string           `json:"schema"`
		Origin origin           `json:"origin"`
		Inputs []inputReference `json:"inputs"`
	}{"gooo/native-cohort-case-inputs/v1", source, references}
	raw, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(directory, "inputs.json"), append(raw, '\n'), 0640)
}
