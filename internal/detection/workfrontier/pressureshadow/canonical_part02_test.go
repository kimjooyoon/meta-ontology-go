package pressureshadow

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCanonicalReplayUsesIndependentDigestLiteral(t *testing.T) {
	input := s1Input()
	data, err := CanonicalInputBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	if CanonicalInputDigest(input) != "sha256:5903d8e3b48faa88e77064ab394c5265cb395494fed8cac5be47a727d13da825" {
		t.Fatalf("digest = %q", CanonicalInputDigest(input))
	}
	input.Selector.Pressures[0], input.Selector.Pressures[1] = input.Selector.Pressures[1], input.Selector.Pressures[0]
	input.Selector.Paths[0], input.Selector.Paths[1] = input.Selector.Paths[1], input.Selector.Paths[0]
	input.PathCoverage[0], input.PathCoverage[1] = input.PathCoverage[1], input.PathCoverage[0]
	for index := range input.PathCoverage {
		coverage := &input.PathCoverage[index].Coverage
		coverage.PressureRecords[0], coverage.PressureRecords[1] = coverage.PressureRecords[1], coverage.PressureRecords[0]
		coverage.RequiredPressureIDs = []string{"pressure/a", "pressure/b"}
	}
	permuted, err := CanonicalInputBytes(input)
	if err != nil || !bytes.Equal(data, permuted) {
		t.Fatalf("permutation changed canonical bytes: %v", err)
	}
	replayedWire, _ := json.Marshal(input)
	replayed, err := DecodeInput(replayedWire)
	if err != nil || CanonicalInputDigest(replayed) != CanonicalInputDigest(input) {
		t.Fatalf("root replay changed canonical digest: %v", err)
	}
	input.PathCoverage[0].PolicyDigest = ""
	input.PathCoverage[0].Coverage.RequiredPressureIDs = []string{"pressure/other"}
	if _, err := CanonicalInputBytes(input); err != nil {
		t.Fatalf("canonical data rejected: %v", err)
	}
	input.PathCoverage = nil
	if _, err := CanonicalInputBytes(input); err != nil {
		t.Fatalf("zero-row canonical data rejected: %v", err)
	}
}
func addRootField(data []byte, field string) []byte {
	return append(append(append([]byte{}, data[:len(data)-1]...), ','), append([]byte(field), '}')...)
}
