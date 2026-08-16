package pressureshadow

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

	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func (input *Input) UnmarshalJSON(data []byte) error {
	type wire Input
	var decoded wire
	if err := decodeStrictObject(data, &decoded); err != nil {
		return err
	}
	if err := validateSyntax(Input(decoded)); err != nil {
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

func decodeStrictObject(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func validateSyntax(input Input) error {
	if input.Schema != SchemaVersion || input.Selector.SchemaVersion != workfrontier.SchemaVersion {
		return fmt.Errorf("invalid schema")
	}
	seenPaths := make(map[string]struct{}, len(input.Selector.Paths))
	for _, pressure := range input.Selector.Pressures {
		if !validID(pressureID(pressure)) {
			return fmt.Errorf("invalid pressure ID")
		}
	}
	for _, path := range input.Selector.Paths {
		id := pathID(path)
		if !validID(id) || !validIDs(path.RequiredPressureIDs) {
			return fmt.Errorf("invalid path ID")
		}
		if _, exists := seenPaths[id]; exists {
			return fmt.Errorf("duplicate selector path ID %q", id)
		}
		seenPaths[id] = struct{}{}
	}
	seenRows := make(map[string]struct{}, len(input.PathCoverage))
	for _, row := range input.PathCoverage {
		if !validID(row.PathID) {
			return fmt.Errorf("invalid path coverage ID")
		}
		if _, exists := seenRows[row.PathID]; exists {
			return fmt.Errorf("duplicate path coverage ID %q", row.PathID)
		}
		seenRows[row.PathID] = struct{}{}
		if _, err := pressurecoverage.CanonicalInputBytes(row.Coverage); err != nil {
			return fmt.Errorf("invalid pressure coverage: %w", err)
		}
	}
	return nil
}

func CanonicalInputBytes(input Input) ([]byte, error) {
	if err := validateSyntax(input); err != nil {
		return nil, err
	}
	rows := append([]PathCoverage{}, input.PathCoverage...)
	sort.Slice(rows, func(left, right int) bool { return rows[left].PathID < rows[right].PathID })
	type canonicalRow struct {
		PathID         string          `json:"path_id"`
		SnapshotDigest string          `json:"snapshot_digest"`
		PolicyDigest   string          `json:"policy_digest"`
		RegistryDigest string          `json:"registry_digest"`
		Coverage       json.RawMessage `json:"coverage"`
	}
	canonicalRows := make([]canonicalRow, 0, len(rows))
	for _, row := range rows {
		coverage, err := pressurecoverage.CanonicalInputBytes(row.Coverage)
		if err != nil {
			return nil, err
		}
		canonicalRows = append(canonicalRows, canonicalRow{
			PathID: row.PathID, SnapshotDigest: row.SnapshotDigest,
			PolicyDigest: row.PolicyDigest, RegistryDigest: row.RegistryDigest, Coverage: coverage,
		})
	}
	return json.Marshal(struct {
		Schema       string             `json:"schema"`
		Selector     workfrontier.Input `json:"selector"`
		PathCoverage []canonicalRow     `json:"path_coverage"`
	}{input.Schema, canonicalSelector(input.Selector), canonicalRows})
}

func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}

func canonicalSelector(input workfrontier.Input) workfrontier.Input {
	input.Pressures = append([]workfrontier.Pressure{}, input.Pressures...)
	input.States = append([]workfrontier.ObligationState{}, input.States...)
	input.Paths = append([]workfrontier.RepairPath{}, input.Paths...)
	sort.Slice(input.Pressures, func(left, right int) bool {
		return pressureID(input.Pressures[left]) < pressureID(input.Pressures[right])
	})
	sort.Slice(input.States, func(left, right int) bool {
		return input.States[left].ObligationID < input.States[right].ObligationID
	})
	for index := range input.Paths {
		path := &input.Paths[index]
		path.PrerequisiteObligationIDs = sortedStrings(path.PrerequisiteObligationIDs)
		path.ReadSet, path.WriteSet = sortedStrings(path.ReadSet), sortedStrings(path.WriteSet)
		path.RequiredPressureIDs = sortedStrings(path.RequiredPressureIDs)
	}
	sort.Slice(input.Paths, func(left, right int) bool {
		return pathID(input.Paths[left]) < pathID(input.Paths[right])
	})
	return input
}

func validIDs(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !validID(value) {
			return false
		}
		if _, exists := seen[value]; exists {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func validID(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func sortedStrings(values []string) []string {
	result := append([]string{}, values...)
	sort.Strings(result)
	return result
}

func pressureID(pressure workfrontier.Pressure) string {
	if pressure.StableID != "" {
		return pressure.StableID
	}
	return pressure.ID
}

func pathID(path workfrontier.RepairPath) string {
	if path.StableID != "" {
		return path.StableID
	}
	return path.ID
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

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
	switch delimiter {
	case '{':
		return scanJSONObject(decoder)
	case '[':
		return scanJSONArray(decoder)
	default:
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
}

func scanJSONObject(decoder *json.Decoder) error {
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
