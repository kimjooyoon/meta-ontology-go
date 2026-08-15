package bindingcoverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
)

func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		output := Output{SchemaVersion: SchemaVersion, InputBytes: uint64(len(data))}
		if schema, ok := schemaFromJSON(data); ok && schema != SchemaVersion {
			output.SchemaVersion = schema
			return seal(output, DecisionUnknown, ReasonUnknownSchema)
		}
		return seal(output, DecisionUnknown, ReasonMissingInput)
	}
	return Observe(input)
}

func ObserveJSON(data []byte) Output { return ClassifyJSON(data) }

func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode binding coverage input: %w", err)
	}
	return normalizeInput(input), nil
}

func Decode(data []byte) (Input, error) { return DecodeJSON(data) }

func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode binding coverage input: %w", err)
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

func EncodeJSON(output Output) ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = output.StableDigest()
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("encode binding coverage output: %w", err)
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
	input.RequiredBindings = copyBindings(input.RequiredBindings)
	input.Partitions = copyPartitions(input.Partitions)
	sort.SliceStable(input.RequiredBindings, func(i, j int) bool {
		return input.RequiredBindings[i].BindingID < input.RequiredBindings[j].BindingID
	})
	sort.SliceStable(input.Partitions, func(i, j int) bool {
		return input.Partitions[i].PartitionID < input.Partitions[j].PartitionID
	})
	return input
}

func copyBindings(values []RequiredBinding) []RequiredBinding {
	if values == nil {
		return nil
	}
	return append(make([]RequiredBinding, 0, len(values)), values...)
}

func copyPartitions(values []Partition) []Partition {
	if values == nil {
		return nil
	}
	return append(make([]Partition, 0, len(values)), values...)
}

func normalizeOutput(output Output) Output {
	output.MissingMatchBindingIDs = sortedStrings(output.MissingMatchBindingIDs)
	output.MissingMismatchBindingIDs = sortedStrings(output.MissingMismatchBindingIDs)
	return output
}

func sortedStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	return requireEOF(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	if delimiter == '{' {
		return scanJSONObject(decoder)
	}
	if delimiter == '[' {
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

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}

func requireInputFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("binding coverage input must be an object")
	}
	for _, name := range []string{"schema_version", "contract_id", "snapshot_digest", "required_bindings", "partitions"} {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("binding coverage input missing %q", name)
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
