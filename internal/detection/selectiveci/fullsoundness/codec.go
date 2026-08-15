package fullsoundness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

func DecodeJSON(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	if err := requireInputFields(data); err != nil {
		return Input{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var input Input
	if err := decoder.Decode(&input); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode full soundness input: %w", err)
	}
	return normalizeInput(input), nil
}

func ClassifyJSON(data []byte) Output {
	input, err := DecodeJSON(data)
	if err != nil {
		return seal(Output{SchemaVersion: SchemaVersion}, DecisionUnknown, ReasonFullSuiteRequired)
	}
	return Evaluate(input)
}

func EncodeInputJSON(input Input) ([]byte, error) {
	data, err := json.Marshal(normalizeInput(input))
	if err != nil {
		return nil, fmt.Errorf("encode full soundness input: %w", err)
	}
	return append(data, '\n'), nil
}

func (input Input) CanonicalJSON() ([]byte, error) { return json.Marshal(normalizeInput(input)) }

func EncodeJSON(output Output) ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = output.StableDigest()
	data, err := json.Marshal(output)
	if err != nil {
		return nil, fmt.Errorf("encode full soundness output: %w", err)
	}
	return append(data, '\n'), nil
}

func (output Output) CanonicalJSON() ([]byte, error) {
	output = normalizeOutput(output)
	output.CanonicalDigest = ""
	return json.Marshal(output)
}

func (output Output) StableDigest() string {
	data, err := output.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
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
	switch delimiter {
	case '{':
		return scanJSONObject(decoder)
	case '[':
		return scanJSONArray(decoder)
	default:
		return fmt.Errorf("unexpected JSON delimiter")
	}
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
		return fmt.Errorf("full soundness input must be an object")
	}
	for _, name := range inputFieldNames {
		if _, exists := fields[name]; !exists {
			return fmt.Errorf("full soundness input missing %q", name)
		}
	}
	return nil
}

var inputFieldNames = []string{
	"schema_version", "snapshot_digest", "policy_digest", "registry_digest", "selection_digest",
	"toolchain_digest", "runner_digest", "obligations", "commands", "impacted_obligation_ids",
	"selected_command_ids", "selection_receipt", "full_outcomes", "selected_outcomes",
	"full_resource_receipts", "selected_resource_receipts", "execution_authorized", "ci_authorized",
}
