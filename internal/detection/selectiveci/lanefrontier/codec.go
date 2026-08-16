package lanefrontier

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

// ClassifyJSON strictly decodes one input value. Malformed or incomplete
// input becomes the same fail-closed UNKNOWN result as missing facts.
func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		if schema, ok := schemaFromJSON(data); ok && schema != SchemaVersion {
			return seal(Output{SchemaVersion: schema}, DecisionUnknown, ReasonUnknownSchema)
		}
		return seal(Output{SchemaVersion: SchemaVersion}, DecisionUnknown, ReasonMissingInput)
	}
	return Classify(input)
}

// DecodeJSON accepts exactly one versioned input object with no duplicate or
// unknown fields. Semantic eligibility is evaluated separately by Classify.
func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode lane frontier input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode lane frontier input: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Input{}, fmt.Errorf("decode lane frontier input: trailing data")
	}
	return normalizeInput(input), nil
}

func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

// EncodeInputJSON emits deterministic input JSON. Classify remains the
// authority for deciding whether the facts are complete and eligible.
func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode lane frontier input: %w", err)
	}
	return append(data, '\n'), nil
}

func (input Input) CanonicalJSON() ([]byte, error) {
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

// EncodeJSON emits a sealed output with its digest bound to the canonical
// representation whose canonical_digest field is empty.
func EncodeJSON(output Output) ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = output.StableDigest()
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("encode lane frontier output: %w", err)
	}
	return append(data, '\n'), nil
}

func (output Output) CanonicalJSON() ([]byte, error) {
	canonical := normalizeOutput(output)
	canonical.CanonicalDigest = ""
	return json.Marshal(canonical)
}

func (output Output) Canonical() string {
	data, err := output.CanonicalJSON()
	if err != nil {
		return ""
	}
	return string(data)
}

func (output Output) StableDigest() string {
	return digestBytes([]byte(output.Canonical()))
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func normalizeInput(input Input) Input {
	input.OwnedPathPrefixes = sortedCopy(input.OwnedPathPrefixes)
	input.ChangedPaths = sortedUnique(input.ChangedPaths)
	return input
}

func normalizeOutput(output Output) Output {
	output.OwnedPathPrefixes = sortedUnique(output.OwnedPathPrefixes)
	output.ChangedPaths = sortedUnique(output.ChangedPaths)
	return output
}

func sortedUnique(values []string) []string {
	values = sortedCopy(values)
	if len(values) < 2 {
		return values
	}
	result := values[:1]
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func sortedCopy(values []string) []string {
	if values == nil {
		return []string{}
	}
	copyValues := append([]string{}, values...)
	sort.Strings(copyValues)
	return copyValues
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
		return scanJSONObject(decoder)
	}
	if delim == '[' {
		return scanJSONArray(decoder)
	}
	return fmt.Errorf("unexpected JSON delimiter")
}

func scanJSONObject(decoder *json.Decoder) error {
	seen := map[string]struct{}{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := token.(string)
		if !ok {
			return fmt.Errorf("object key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate object field %q", name)
		}
		seen[name] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}

func scanJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}

func requireInputFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("lane frontier input must be an object")
	}
	for _, name := range []string{
		"schema_version", "registry_digest", "base_sha", "lane_head_sha",
		"lane_id", "registered_branch", "owned_path_prefixes", "changed_paths",
		"ahead_count", "behind_count", "open_pr_count", "active_lease_count",
	} {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("lane frontier input missing %q", name)
		}
	}
	return nil
}

func schemaFromJSON(data []byte) (string, bool) {
	if rejectDuplicateKeys(data) != nil {
		return "", false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return "", false
	}
	var schema string
	if err := json.Unmarshal(fields["schema_version"], &schema); err != nil {
		return "", false
	}
	return schema, true
}
