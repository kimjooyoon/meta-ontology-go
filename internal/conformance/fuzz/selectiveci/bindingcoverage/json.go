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

func canonicalInputJSON(input Input) ([]byte, error) {
	normalized := input
	normalized.Precedence = append([]Precedence{}, input.Precedence...)
	normalized.RequiredBindings = append([]Binding{}, input.RequiredBindings...)
	normalized.Partitions = append([]Partition{}, input.Partitions...)
	sort.Slice(normalized.Precedence, func(i, j int) bool {
		return normalized.Precedence[i].Rank < normalized.Precedence[j].Rank
	})
	sort.Slice(normalized.RequiredBindings, func(i, j int) bool {
		return normalized.RequiredBindings[i].ID < normalized.RequiredBindings[j].ID
	})
	sort.Slice(normalized.Partitions, func(i, j int) bool {
		left, right := normalized.Partitions[i], normalized.Partitions[j]
		if left.BindingID != right.BindingID {
			return left.BindingID < right.BindingID
		}
		return left.Polarity < right.Polarity
	})
	return json.Marshal(normalized)
}

func digestVector(vector Vector) string {
	data, _ := json.Marshal(vector)
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func decodeStrict(data []byte, target any) error {
	if err := rejectDuplicateKeys(data); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
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
	seen := make(map[string]bool)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := keyToken.(string)
		if !ok || seen[key] {
			return fmt.Errorf("duplicate or invalid JSON key")
		}
		seen[key] = true
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}
