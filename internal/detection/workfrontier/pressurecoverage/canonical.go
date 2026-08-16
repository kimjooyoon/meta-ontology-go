package pressurecoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"unicode"
)

func CanonicalInputBytes(input Input) ([]byte, error) { return json.Marshal(normalizeInput(input)) }

func CanonicalInputDigest(input Input) string {
	data, err := CanonicalInputBytes(input)
	if err != nil {
		return digestBytes([]byte("canonical-input-error:" + err.Error()))
	}
	return digestBytes(data)
}

func CanonicalOutputBytes(output Output) ([]byte, error) {
	output.SelectedIDs, output.UnselectedIDs, output.UnknownIDs = sortedUnique(output.SelectedIDs), sortedUnique(output.UnselectedIDs), sortedUnique(output.UnknownIDs)
	output.OutputDigest, output.ReplayDigest = "", ""
	return json.Marshal(output)
}

func CanonicalOutputDigest(output Output) string {
	data, err := CanonicalOutputBytes(output)
	if err != nil {
		return digestBytes([]byte("canonical-output-error:" + err.Error()))
	}
	return digestBytes(data)
}

func ReplayDigest(inputDigest, outputDigest string) string {
	return digestBytes([]byte(inputDigest + "\x00" + outputDigest))
}

func normalizeInput(input Input) Input {
	input.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	input.RequiredPressureIDs, input.FinitePathIDs = sortedCopy(input.RequiredPressureIDs), sortedCopy(input.FinitePathIDs)
	input.GuardIDs, input.PresentationRoot = sortedCopy(input.GuardIDs), ""
	for i := range input.PressureRecords {
		input.PressureRecords[i].DisplayName, input.PressureRecords[i].PresentationMetadata = "", nil
	}
	sort.Slice(input.PressureRecords, func(i, j int) bool {
		return pressureKey(input.PressureRecords[i]) < pressureKey(input.PressureRecords[j])
	})
	return input
}

func pressureKey(record PressureRecord) string {
	return strings.Join([]string{record.PressureID, record.CategoryID, record.IndependenceGroupID, record.ApplicabilityRuleID}, "\x00")
}

func sortedUnique(values []string) []string {
	if values == nil {
		return []string{}
	}
	result := sortedCopy(values)
	if len(result) < 2 {
		return result
	}
	out := result[:1]
	for _, value := range result[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(value[7:])
	if err != nil {
		return false
	}
	for _, value := range decoded {
		if value != 0 {
			return true
		}
	}
	return false
}

func validID(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
