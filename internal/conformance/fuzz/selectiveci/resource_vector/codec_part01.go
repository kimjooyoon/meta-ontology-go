package resourcevector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode resource-vector input: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Input{}, fmt.Errorf("decode resource-vector input: trailing JSON value")
		}
		return Input{}, fmt.Errorf("decode resource-vector input: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return Input{}, fmt.Errorf("decode resource-vector input: expected object")
	}
	for _, name := range []string{"schema", "fixture_id", "root", "commands", "paths", "affected_stable_ids", "selected_command_ids", "full_command_ids", "ceilings"} {
		if !present(fields, name) {
			return Input{}, fmt.Errorf("decode resource-vector input: missing %q", name)
		}
	}
	return input, nil
}
func EncodeInputJSON(input Input) ([]byte, error) {
	normalized := canonicalInput(input)
	data, err := json.MarshalIndent(struct {
		Schema             string           `json:"schema"`
		FixtureID          string           `json:"fixture_id"`
		Root               string           `json:"root"`
		Commands           []CommandRecord  `json:"commands"`
		Paths              []PathRecord     `json:"paths"`
		AffectedStableIDs  []string         `json:"affected_stable_ids"`
		SelectedCommandIDs []string         `json:"selected_command_ids"`
		FullCommandIDs     []string         `json:"full_command_ids"`
		Ceilings           ResourceCeilings `json:"ceilings"`
	}{
		Schema: normalized.Schema, FixtureID: normalized.FixtureID, Root: input.Root,
		Commands: normalized.Commands, Paths: normalized.Paths,
		AffectedStableIDs:  normalized.AffectedStableIDs,
		SelectedCommandIDs: normalized.SelectedCommandIDs, FullCommandIDs: normalized.FullCommandIDs,
		Ceilings: normalized.Ceilings,
	}, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
func CanonicalInputBytes(input Input) ([]byte, error) {
	return json.Marshal(canonicalInput(input))
}
func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}
