package selectiveci

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/impactgraph"
)

func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode selective-ci input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode selective-ci input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Input{}, fmt.Errorf("decode selective-ci input: trailing data")
	}
	if err := input.Validate(); err != nil {
		return Input{}, err
	}
	return normalizeInput(input), nil
}

func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

func EncodeJSON(input Input) ([]byte, error) {
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}

func (input Input) CanonicalJSON() ([]byte, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(normalizeInput(input))
}

func (input Input) Canonical() string {
	data, err := input.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}

func (input Input) Digest() string { return digestBytes([]byte(input.Canonical())) }

func EncodePlanJSON(result PlanResult) ([]byte, error) {
	result = sealResult(result)
	canonical, err := json.Marshal(normalizeResult(result))
	if err != nil {
		return nil, err
	}
	return append(canonical, '\n'), nil
}

func (result PlanResult) CanonicalJSON() ([]byte, error) {
	copy := normalizeResult(result)
	copy.CanonicalDigest = ""
	canonical, err := json.Marshal(copy)
	if err != nil {
		return nil, err
	}
	return canonical, nil
}

func (result PlanResult) Canonical() string {
	data, err := result.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}

func (result PlanResult) StableDigest() string {
	return digestBytes([]byte(result.Canonical()))
}

func sealResult(result PlanResult) PlanResult {
	result = normalizeResult(result)
	result.CanonicalDigest = result.StableDigest()
	result.Digest = result.CanonicalDigest
	return result
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeInput(input Input) Input {
	input.Base = normalizeManifest(input.Base)
	input.Head = normalizeManifest(input.Head)
	input.Registry = normalizeRegistry(input.Registry)
	input.Receipts = append([]Receipt{}, input.Receipts...)
	input.ProvenancePaths = append([]ProvenancePath{}, input.ProvenancePaths...)
	sort.Slice(input.Receipts, func(i, j int) bool { return input.Receipts[i].CommandID < input.Receipts[j].CommandID })
	sort.Slice(input.ProvenancePaths, func(i, j int) bool { return input.ProvenancePaths[i].CommandID < input.ProvenancePaths[j].CommandID })
	for i := range input.ProvenancePaths {
		if normalized, err := input.ProvenancePaths[i].Path.Normalized(); err == nil {
			input.ProvenancePaths[i].Path = normalized
		}
	}
	return input
}

func normalizeManifest(manifest SnapshotManifest) SnapshotManifest {
	manifest.Files = append([]SnapshotFile{}, manifest.Files...)
	for i := range manifest.Files {
		manifest.Files[i].SemanticIDs = sortedCopy(manifest.Files[i].SemanticIDs)
	}
	sort.Slice(manifest.Files, func(i, j int) bool { return manifest.Files[i].Path < manifest.Files[j].Path })
	return manifest
}

func normalizeRegistry(registry Registry) Registry {
	registry.Nodes = append([]impactgraph.Node{}, registry.Nodes...)
	registry.DependencyEdges = append([]DependencyEdge{}, registry.DependencyEdges...)
	registry.Obligations = append([]ObligationBinding{}, registry.Obligations...)
	registry.Commands = append([]Command{}, registry.Commands...)
	registry.GlobalGuardCommands = append([]Command{}, registry.GlobalGuardCommands...)
	sort.Slice(registry.Nodes, func(i, j int) bool { return registry.Nodes[i].ID < registry.Nodes[j].ID })
	sort.Slice(registry.DependencyEdges, func(i, j int) bool {
		return edgeKey(registry.DependencyEdges[i]) < edgeKey(registry.DependencyEdges[j])
	})
	sort.Slice(registry.Obligations, func(i, j int) bool { return registry.Obligations[i].ID < registry.Obligations[j].ID })
	sort.Slice(registry.Commands, func(i, j int) bool { return registry.Commands[i].ID < registry.Commands[j].ID })
	sort.Slice(registry.GlobalGuardCommands, func(i, j int) bool { return registry.GlobalGuardCommands[i].ID < registry.GlobalGuardCommands[j].ID })
	for i := range registry.Obligations {
		registry.Obligations[i].CommandIDs = sortedCopy(registry.Obligations[i].CommandIDs)
	}
	return registry
}

func normalizeResult(result PlanResult) PlanResult {
	result.ChangedSemanticIDs = sortedUnique(result.ChangedSemanticIDs)
	result.SelectedCommandIDs = sortedUnique(result.SelectedCommandIDs)
	result.SelectedGuardCommandIDs = sortedUnique(result.SelectedGuardCommandIDs)
	result.SelectedWorkIDs = sortedUnique(result.SelectedWorkIDs)
	result.ResourceReceiptDigests = sortedUnique(result.ResourceReceiptDigests)
	result.ProvenancePathIDs = sortedUnique(result.ProvenancePathIDs)
	return result
}

func sortedCopy(values []string) []string {
	if values == nil {
		return nil
	}
	copy := append([]string{}, values...)
	sort.Strings(copy)
	return copy
}

func sortedUnique(values []string) []string {
	values = sortedCopy(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func edgeKey(edge DependencyEdge) string {
	return edge.From + "\x00" + string(edge.Kind) + "\x00" + edge.To
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
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
	if delim == '{' {
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name := key.(string)
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate object field %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delim == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	}
	_, err = decoder.Token()
	return err
}

func requireInputFields(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil || root == nil {
		return failure(ReasonInvalidInput, "input must be an object")
	}
	if err := requireFields(root, "schema_version", "base", "head", "registry", "cpu_capacity", "resource_receipts", "provenance_paths"); err != nil {
		return err
	}
	for _, name := range []string{"base", "head"} {
		fields, err := objectFields(root[name])
		if err != nil {
			return err
		}
		if err := requireFields(fields, "schema_version", "snapshot_digest", "files"); err != nil {
			return err
		}
	}
	registry, err := objectFields(root["registry"])
	if err != nil {
		return err
	}
	if err := requireFields(registry, "schema_version", "registry_digest", "policy_digest", "nodes", "dependency_edges", "obligations", "commands", "global_guard_commands"); err != nil {
		return err
	}
	if err := requireArrayObjectFields(root["resource_receipts"], "command_id", "snapshot_digest", "envelope"); err != nil {
		return err
	}
	return requireArrayObjectFields(root["provenance_paths"], "command_id", "path", "requirement")
}

func objectFields(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, failure(ReasonInvalidInput, "required field must be an object")
	}
	return fields, nil
}

func requireFields(fields map[string]json.RawMessage, names ...string) error {
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return failure(ReasonInvalidInput, "missing required field "+name)
		}
	}
	return nil
}

func requireArrayObjectFields(raw json.RawMessage, names ...string) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return failure(ReasonInvalidInput, "required field must be an array")
	}
	for _, value := range values {
		fields, err := objectFields(value)
		if err != nil {
			return err
		}
		if err := requireFields(fields, names...); err != nil {
			return err
		}
	}
	return nil
}
