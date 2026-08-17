package impactcoverage

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
)

type inputWire struct {
	Schema string                `json:"schema"`
	Base   *selectiveci.Snapshot `json:"base"`
	Head   *selectiveci.Snapshot `json:"head"`
}

func (input Input) CanonicalJSON() ([]byte, error) {
	if input.Schema != SchemaV1 {
		return nil, fmt.Errorf("impact coverage schema %q is invalid", input.Schema)
	}
	if input.Base == nil || input.Head == nil {
		return nil, fmt.Errorf("base and head snapshots are required")
	}
	if err := input.Base.Validate(); err != nil {
		return nil, fmt.Errorf("base snapshot: %w", err)
	}
	if err := input.Head.Validate(); err != nil {
		return nil, fmt.Errorf("head snapshot: %w", err)
	}
	return json.Marshal(inputWire{Schema: input.Schema, Base: input.Base, Head: input.Head})
}

func (input Input) Digest() string {
	data, err := input.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func (input Input) MarshalJSON() ([]byte, error) { return input.CanonicalJSON() }

func (input *Input) UnmarshalJSON(data []byte) error {
	decoded, err := DecodeInput(data)
	if err != nil {
		return err
	}
	*input = decoded
	return nil
}

func DecodeInput(data []byte) (Input, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire inputWire
	if err := decoder.Decode(&wire); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Input{}, fmt.Errorf("decode impact coverage input: %w", err)
	}
	input := Input{Schema: wire.Schema, Base: wire.Base, Head: wire.Head}
	canonical, err := input.CanonicalJSON()
	if err != nil {
		return Input{}, err
	}
	if !bytes.Equal(canonical, data) {
		return Input{}, fmt.Errorf("impact coverage input is not canonical")
	}
	return input, nil
}

func ObserveJSON(data []byte) Result {
	input, err := DecodeInput(data)
	if err != nil {
		result := Result{Schema: SchemaV1, ChangedStableIDs: []string{}, UncoveredPaths: []string{}}
		return seal(result, DecisionUnknown, ReasonInvalidSnapshot)
	}
	return Observe(input)
}

func EvaluateJSON(data []byte) Result { return ObserveJSON(data) }

func Decode(data []byte) (Input, error) { return DecodeInput(data) }

func EncodeInputJSON(input Input) ([]byte, error) { return input.CanonicalJSON() }

func (result Result) CanonicalJSON() ([]byte, error) {
	normalized, err := normalizeResult(result)
	if err != nil {
		return nil, err
	}
	normalized.OutputDigest = ""
	return json.Marshal(normalized)
}

func (result Result) StableDigest() string {
	data, err := result.CanonicalJSON()
	if err != nil {
		return ""
	}
	return digestBytes(data)
}

func EncodeJSON(result Result) ([]byte, error) {
	normalized, err := normalizeResult(result)
	if err != nil {
		return nil, err
	}
	digest := normalized.StableDigest()
	if normalized.OutputDigest != "" && normalized.OutputDigest != digest {
		return nil, fmt.Errorf("impact coverage output digest mismatch")
	}
	normalized.OutputDigest = digest
	return json.Marshal(normalized)
}

func DecodeJSON(data []byte) (Result, error) {
	if err := rejectDuplicateKeys(data); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var result Result
	if err := decoder.Decode(&result); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	if err := requireEOF(decoder); err != nil {
		return Result{}, fmt.Errorf("decode impact coverage output: %w", err)
	}
	normalized, err := normalizeResult(result)
	if err != nil {
		return Result{}, err
	}
	encoded, err := EncodeJSON(normalized)
	if err != nil {
		return Result{}, err
	}
	if !bytes.Equal(encoded, data) {
		return Result{}, fmt.Errorf("impact coverage output is not canonical")
	}
	return normalized, nil
}

func normalizeResult(result Result) (Result, error) {
	if result.Schema != SchemaV1 {
		return Result{}, fmt.Errorf("impact coverage schema %q is invalid", result.Schema)
	}
	if result.Decision != DecisionExact && result.Decision != DecisionUnknown {
		return Result{}, fmt.Errorf("impact coverage decision %q is invalid", result.Decision)
	}
	if !validReason(result.Decision, result.Reason) {
		return Result{}, fmt.Errorf("impact coverage reason %q is invalid", result.Reason)
	}
	if result.FullSuiteRequired != (result.Decision == DecisionUnknown) {
		return Result{}, fmt.Errorf("impact coverage full-suite flag is inconsistent")
	}
	result.ChangedStableIDs = sortedUnique(result.ChangedStableIDs)
	result.UncoveredPaths = sortedUnique(result.UncoveredPaths)
	if result.Decision == DecisionUnknown && len(result.ChangedStableIDs) != 0 {
		return Result{}, fmt.Errorf("UNKNOWN result cannot contain changed stable IDs")
	}
	if result.ChangedStableIDs == nil {
		result.ChangedStableIDs = []string{}
	}
	if result.UncoveredPaths == nil {
		result.UncoveredPaths = []string{}
	}
	return result, nil
}

func validReason(decision Decision, reason Reason) bool {
	if decision == DecisionExact {
		return reason == ReasonComplete || reason == ReasonNoChange
	}
	return reason == ReasonMissingBinding || reason == ReasonAuthorityDrift || reason == ReasonInvalidSnapshot
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
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
	if delimiter == '[' {
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	if delimiter != '{' {
		return fmt.Errorf("unexpected JSON delimiter %q", delimiter)
	}
	seen := make(map[string]struct{})
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return err
		}
		name, ok := key.(string)
		if !ok {
			return fmt.Errorf("JSON object key is not a string")
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate JSON field %q", name)
		}
		seen[name] = struct{}{}
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}
