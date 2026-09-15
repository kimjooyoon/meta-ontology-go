package pressurecoverage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

func a2Input() Input {
	return Input{
		Schema:                  SchemaVersion,
		AuthoritySnapshotDigest: a2Snapshot,
		PolicyDigest:            a2Policy,
		RegistryDigest:          a2Registry,
		ToolchainOptionsDigest:  a2Toolchain,
		RequestedK:              2,
		MinimumIndependent:      2,
		PressureRecords: []PressureRecord{
			{"p-c", "category-c", "group-c", "rule-1"},
			{"p-a", "category-a", "group-a", "rule-1"},
			{"p-b", "category-b", "group-b", "rule-1"},
		},
		RequiredPressureIDs: []string{"p-c", "p-a", "p-b"},
	}
}

// testBind independently rebinds mutated fixtures with raw SHA-256 construction.
func testBind(input *Input) {
	unsigned := *input
	unsigned.AuthoritySnapshotDigest = ""
	unsigned.PolicyDigest = ""
	unsigned.RegistryDigest = ""
	unsigned.ToolchainOptionsDigest = ""
	unsigned.PressureRecords = append([]PressureRecord(nil), unsigned.PressureRecords...)
	unsigned.RequiredPressureIDs = append([]string(nil), unsigned.RequiredPressureIDs...)
	sort.Slice(unsigned.PressureRecords, func(left, right int) bool {
		return unsigned.PressureRecords[left].PressureID < unsigned.PressureRecords[right].PressureID
	})
	sort.Strings(unsigned.RequiredPressureIDs)
	data, _ := json.Marshal(unsigned)
	inputDigest := testDigest(data)
	input.AuthoritySnapshotDigest = testRoleDigest("authority-snapshot", inputDigest)
	input.PolicyDigest = testRoleDigest("policy", inputDigest)
	input.RegistryDigest = testRoleDigest("registry", inputDigest)
	input.ToolchainOptionsDigest = testRoleDigest("toolchain-options", inputDigest)
}
func testRoleDigest(role, inputDigest string) string {
	return testDigest([]byte(role + "\x00" + inputDigest))
}
func testDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
