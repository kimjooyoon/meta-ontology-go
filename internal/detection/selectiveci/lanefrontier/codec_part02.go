package lanefrontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

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
