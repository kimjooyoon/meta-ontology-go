package resourcevector

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
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

func EncodeOutputJSON(output Output) ([]byte, error) {
	data, err := json.MarshalIndent(canonicalOutput(output), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func CanonicalOutputDigest(output Output) string {
	data, err := json.Marshal(canonicalOutput(output))
	if err != nil {
		return digestBytes([]byte("canonical-output-error:" + err.Error()))
	}
	return digestBytes(data)
}

func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}

type canonicalOutputView struct {
	Schema            string   `json:"schema"`
	FixtureID         string   `json:"fixture_id"`
	InputDigest       string   `json:"input_digest"`
	Selected          *Vector  `json:"selected,omitempty"`
	Full              *Vector  `json:"full,omitempty"`
	Decision          Decision `json:"decision"`
	Reason            Reason   `json:"reason"`
	LimitFailures     []string `json:"limit_failures"`
	FullSuiteRequired bool     `json:"full_suite_required"`
	ProofValid        bool     `json:"proof_valid"`
}

func canonicalOutput(output Output) canonicalOutputView {
	return canonicalOutputView{
		Schema: output.Schema, FixtureID: output.FixtureID, InputDigest: output.InputDigest,
		Selected: output.Selected, Full: output.Full, Decision: output.Decision, Reason: output.Reason,
		LimitFailures: sortedStrings(output.LimitFailures), FullSuiteRequired: output.FullSuiteRequired,
		ProofValid: output.ProofValid,
	}
}

func canonicalInput(input Input) canonicalInputView {
	commands := append([]CommandRecord(nil), input.Commands...)
	for index := range commands {
		if path, ok := canonicalRelativePath(input.Root, commands[index].Path); ok {
			commands[index].Path = path
		}
		commands[index].Pressures = append([]PressureRecord(nil), commands[index].Pressures...)
		commands[index].AffectedStableIDs = sortedStrings(commands[index].AffectedStableIDs)
		sort.Slice(commands[index].Pressures, func(left, right int) bool {
			return pressureKey(commands[index].Pressures[left]) < pressureKey(commands[index].Pressures[right])
		})
	}
	sort.Slice(commands, func(left, right int) bool { return commands[left].ID < commands[right].ID })
	paths := append([]PathRecord(nil), input.Paths...)
	for index := range paths {
		if path, ok := canonicalRelativePath(input.Root, paths[index].Path); ok {
			paths[index].Path = path
		}
		paths[index].RecordIDs = sortedStrings(paths[index].RecordIDs)
	}
	sort.Slice(paths, func(left, right int) bool {
		if paths[left].ID != paths[right].ID {
			return paths[left].ID < paths[right].ID
		}
		if paths[left].Path != paths[right].Path {
			return paths[left].Path < paths[right].Path
		}
		return paths[left].CommandID < paths[right].CommandID
	})
	return canonicalInputView{
		Schema: input.Schema, FixtureID: input.FixtureID, Commands: commands, Paths: paths,
		AffectedStableIDs:  sortedStrings(input.AffectedStableIDs),
		SelectedCommandIDs: sortedStrings(input.SelectedCommandIDs), FullCommandIDs: sortedStrings(input.FullCommandIDs),
		Ceilings: input.Ceilings,
	}
}

type canonicalInputView struct {
	Schema             string           `json:"schema"`
	FixtureID          string           `json:"fixture_id"`
	Commands           []CommandRecord  `json:"commands"`
	Paths              []PathRecord     `json:"paths"`
	AffectedStableIDs  []string         `json:"affected_stable_ids"`
	SelectedCommandIDs []string         `json:"selected_command_ids"`
	FullCommandIDs     []string         `json:"full_command_ids"`
	Ceilings           ResourceCeilings `json:"ceilings"`
}

func canonicalRelativePath(root, value string) (string, bool) {
	root = strings.ReplaceAll(root, "\\", "/")
	value = strings.ReplaceAll(value, "\\", "/")
	if root == "" || value == "" || strings.ContainsAny(root+value, "\x00") {
		return "", false
	}
	root = strings.TrimSuffix(root, "/")
	if strings.HasPrefix(value, root+"/") {
		value = strings.TrimPrefix(value, root+"/")
	} else if strings.HasPrefix(value, "/") {
		return "", false
	}
	if value == "" || strings.HasPrefix(value, "/") {
		return "", false
	}
	parts := strings.Split(value, "/")
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return "", false
		}
	}
	return strings.Join(parts, "/"), true
}

func pressureKey(record PressureRecord) string {
	applicable := "false"
	if record.Applicable != nil && *record.Applicable {
		applicable = "true"
	}
	return record.ID + "\x00" + record.IndependenceGroupID + "\x00" + applicable
}

func present(fields map[string]json.RawMessage, name string) bool {
	raw, ok := fields[name]
	return ok && !bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("multiple JSON values")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delim)
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[key]; exists {
			return fmt.Errorf("duplicate JSON key %q", key)
		}
		seen[key] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
