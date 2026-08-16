package pressurecoverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// UnmarshalJSON is the strict public boundary for Input.
func (input *Input) UnmarshalJSON(data []byte) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	type wire Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded wire
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	if err := validateInput(Input(decoded)); err != nil {
		return err
	}
	*input = Input(decoded)
	return nil
}

func DecodeInput(data []byte) (Input, error) {
	var input Input
	if err := json.Unmarshal(data, &input); err != nil {
		return Input{}, err
	}
	return input, nil
}

func CanonicalInputBytes(input Input) ([]byte, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}
	return json.Marshal(normalizeInput(input))
}

func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}

func authorityBindingDigest(input Input, role string) string {
	input.AuthoritySnapshotDigest = ""
	input.PolicyDigest = ""
	input.RegistryDigest = ""
	input.ToolchainOptionsDigest = ""
	return digestBytes([]byte(role + "\x00" + CanonicalInputDigest(input)))
}

func validateInput(input Input) error {
	if input.Schema != SchemaVersion {
		return fmt.Errorf("pressure coverage input has invalid schema")
	}
	seen := make(map[string]struct{}, len(input.RequiredPressureIDs))
	for _, id := range input.RequiredPressureIDs {
		if !validID(id) {
			return fmt.Errorf("invalid required pressure ID %q", id)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate required pressure ID %q", id)
		}
		seen[id] = struct{}{}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if !validID(record.PressureID) || !validID(record.CategoryID) ||
			!optionalID(record.IndependenceGroupID) || !optionalID(record.ApplicabilityRuleID) {
			return fmt.Errorf("invalid pressure record ID %q", record.PressureID)
		}
		if prior, exists := records[record.PressureID]; exists {
			if prior == record {
				return fmt.Errorf("duplicate pressure ID %q", record.PressureID)
			}
			return fmt.Errorf("conflicting pressure ID %q", record.PressureID)
		}
		records[record.PressureID] = record
	}
	return nil
}

func optionalID(value string) bool { return value == "" || validID(value) }

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	if delimiter == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	} else if delimiter == '{' {
		seen := map[string]struct{}{}
		for decoder.More() {
			key, err := decoder.Token()
			if err != nil {
				return err
			}
			name := key.(string)
			if _, exists := seen[name]; exists {
				return fmt.Errorf("duplicate JSON key %q", name)
			}
			seen[name] = struct{}{}
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	} else {
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	_, err = decoder.Token()
	return err
}

func normalizeInput(input Input) Input {
	input.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	input.RequiredPressureIDs = append([]string(nil), input.RequiredPressureIDs...)
	sort.Slice(input.PressureRecords, func(left, right int) bool {
		return pressureKey(input.PressureRecords[left]) < pressureKey(input.PressureRecords[right])
	})
	sort.Strings(input.RequiredPressureIDs)
	return input
}

func pressureKey(record PressureRecord) string {
	return strings.Join([]string{record.PressureID, record.CategoryID,
		record.IndependenceGroupID, record.ApplicabilityRuleID}, "\x00")
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validID(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 ||
		!utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
