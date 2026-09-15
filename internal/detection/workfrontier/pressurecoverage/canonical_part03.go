package pressurecoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

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
