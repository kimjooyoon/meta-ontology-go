package pressurecoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

func CanonicalInputDigest(input Input) string {
	data, _ := json.Marshal(normalizeInput(input))
	return digestBytes(data)
}

func authorityBindingDigest(input Input, role string) string {
	input.AuthoritySnapshotDigest = ""
	input.PolicyDigest = ""
	input.RegistryDigest = ""
	input.ToolchainOptionsDigest = ""
	return digestBytes([]byte(role + "\x00" + CanonicalInputDigest(input)))
}

func boundDigests(input Input) bool {
	return input.AuthoritySnapshotDigest == authorityBindingDigest(input, "authority-snapshot") &&
		input.PolicyDigest == authorityBindingDigest(input, "policy") &&
		input.RegistryDigest == authorityBindingDigest(input, "registry") &&
		input.ToolchainOptionsDigest == authorityBindingDigest(input, "toolchain-options")
}

func CanonicalOutputDigest(output Output) string {
	output.SelectedIDs = sortedUnique(output.SelectedIDs)
	output.UnselectedIDs = sortedUnique(output.UnselectedIDs)
	output.UnknownIDs = sortedUnique(output.UnknownIDs)
	output.OutputDigest, output.ReplayDigest = "", ""
	data, _ := json.Marshal(output)
	return digestBytes(data)
}

func normalizeInput(input Input) Input {
	input.PressureRecords = append([]PressureRecord(nil), input.PressureRecords...)
	input.RequiredPressureIDs = sortedCopy(input.RequiredPressureIDs)
	input.FinitePathIDs = sortedCopy(input.FinitePathIDs)
	input.GuardIDs = sortedCopy(input.GuardIDs)
	sort.Slice(input.PressureRecords, func(left, right int) bool {
		return pressureKey(input.PressureRecords[left]) < pressureKey(input.PressureRecords[right])
	})
	return input
}

func pressureKey(record PressureRecord) string {
	return strings.Join([]string{record.PressureID, record.CategoryID,
		record.IndependenceGroupID, record.ApplicabilityRuleID}, "\x00")
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func sortedUnique(values []string) []string {
	if values == nil {
		return []string{}
	}
	result := sortedCopy(values)
	if len(result) < 2 {
		return result
	}
	unique := result[:1]
	for _, value := range result[1:] {
		if value != unique[len(unique)-1] {
			unique = append(unique, value)
		}
	}
	return unique
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if len(value) != 71 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[7:])
	return err == nil && strings.Trim(value[7:], "0") != ""
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
