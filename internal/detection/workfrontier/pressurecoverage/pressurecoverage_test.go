package pressurecoverage

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCanonicalRoundTripAndBinding(t *testing.T) {
	input := fixture()
	want, err := CanonicalInputBytes(input)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInput(want)
	if err != nil || CanonicalInputDigest(decoded) != CanonicalInputDigest(input) {
		t.Fatalf("round trip = %#v, error = %v", decoded, err)
	}
	var direct Input
	if err := json.Unmarshal(want, &direct); err != nil {
		t.Fatalf("direct json.Unmarshal: %v", err)
	}
	if direct.RequestedK != 21 || direct.Schema != SchemaVersion {
		t.Fatalf("decoded envelope = %#v", direct)
	}
	if input.PolicyDigest != authorityBindingDigest(input, "policy") ||
		input.RegistryDigest != authorityBindingDigest(input, "registry") {
		t.Fatal("authority binding digest mismatch")
	}
}

func TestCanonicalPermutationReplay(t *testing.T) {
	first, err := CanonicalInputBytes(fixture())
	if err != nil {
		t.Fatal(err)
	}
	input := fixture()
	input.RequiredPressureIDs[0], input.RequiredPressureIDs[3] = input.RequiredPressureIDs[3], input.RequiredPressureIDs[0]
	input.PressureRecords[0], input.PressureRecords[3] = input.PressureRecords[3], input.PressureRecords[0]
	second, err := CanonicalInputBytes(input)
	if err != nil || !bytes.Equal(first, second) || CanonicalInputDigest(input) != digestBytes(first) {
		t.Fatalf("canonical replay differs: %s != %s", first, second)
	}
}

func TestCanonicalInputRejectsMalformedAndAmbiguousRecords(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Input)
	}{
		{"internal whitespace", func(input *Input) { input.RequiredPressureIDs[0] = "pressure z" }},
		{"control character", func(input *Input) { input.PressureRecords[0].CategoryID = "category\x00" }},
		{"duplicate record", func(input *Input) {
			input.PressureRecords = append(input.PressureRecords, input.PressureRecords[0])
		}},
		{"conflicting record", func(input *Input) {
			input.PressureRecords = append(input.PressureRecords, PressureRecord{
				"pressure-a", "category-a", "group-new", "rule-1",
			})
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := fixture()
			test.edit(&input)
			if _, err := CanonicalInputBytes(input); err == nil {
				t.Fatal("accepted malformed or ambiguous input")
			}
		})
	}
}

func TestStrictJSONBoundary(t *testing.T) {
	cases := []struct {
		name string
		data string
	}{
		{"unknown top-level field", schemaJSON(`,"display_name":"shown"`)},
		{"unknown nested field", schemaJSON(`,"pressure_records":[{"pressure_id":"p","display_name":"shown"}]`)},
		{"duplicate top-level key", schemaJSON(`,"schema":"` + SchemaVersion + `"`)},
		{"duplicate nested key", schemaJSON(`,"pressure_records":[{"pressure_id":"p","pressure_id":"q"}]`)},
		{"trailing JSON", schemaJSON("") + " " + schemaJSON("")},
		{"invalid schema", `{"schema":"wrong"}`},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if _, err := DecodeInput([]byte(test.data)); err == nil {
				t.Fatal("DecodeInput accepted invalid JSON")
			}
			var input Input
			if err := json.Unmarshal([]byte(test.data), &input); err == nil {
				t.Fatal("json.Unmarshal accepted invalid JSON")
			}
		})
	}
}

func fixture() Input {
	input := Input{
		Schema:             SchemaVersion,
		RequestedK:         21,
		MinimumIndependent: 2,
		PressureRecords: []PressureRecord{
			{"pressure-z", "category-z", "group-z", "rule-1"},
			{"pressure-a", "category-a", "group-a", "rule-1"},
			{"pressure-b", "category-b", "group-b", "rule-1"},
			{"pressure-aa", "category-a", "group-a", "rule-1"},
		},
		RequiredPressureIDs: []string{"pressure-z", "pressure-a", "pressure-b", "pressure-aa"},
	}
	bindDigests(&input)
	return input
}

func bindDigests(input *Input) {
	input.AuthoritySnapshotDigest = authorityBindingDigest(*input, "authority-snapshot")
	input.PolicyDigest = authorityBindingDigest(*input, "policy")
	input.RegistryDigest = authorityBindingDigest(*input, "registry")
	input.ToolchainOptionsDigest = authorityBindingDigest(*input, "toolchain-options")
}

func schemaJSON(suffix string) string {
	return `{"schema":"` + SchemaVersion + `"` + suffix + `}`
}
