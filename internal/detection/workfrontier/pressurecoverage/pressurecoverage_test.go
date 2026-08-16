package pressurecoverage

import (
	"bytes"
	"encoding/json"
	"testing"
)

const (
	expectedSnapshot  = "sha256:84440ba2628e8ce259a82d21115eb08c78a37dee612f2efdea6e3b7bf0f508c7"
	expectedPolicy    = "sha256:80c9b1f9b9f059c43ad73a4f2a46c740f38e41c987d66e5f7dc9203775e81968"
	expectedRegistry  = "sha256:f325d05cef0f2c5fc1a1f03e9ac93c5a7eab96a10fbec009637f197af0f847af"
	expectedToolchain = "sha256:e5e6b048838a03825a54a519e7e4ee56621d72be359da5a84bb8b774cd57ec7e"
)

func TestCanonicalRoundTripAndBinding(t *testing.T) {
	input := fixture()
	bindings := []struct {
		role string
		want string
	}{
		{"authority-snapshot", expectedSnapshot},
		{"policy", expectedPolicy},
		{"registry", expectedRegistry},
		{"toolchain-options", expectedToolchain},
	}
	for _, binding := range bindings {
		if got := authorityBindingDigest(input, binding.role); got != binding.want {
			t.Fatalf("%s binding = %s, want %s", binding.role, got, binding.want)
		}
	}
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
}

func TestBindingSurfaceMutations(t *testing.T) {
	bindings := []struct {
		role string
		want string
	}{
		{"authority-snapshot", expectedSnapshot},
		{"policy", expectedPolicy},
		{"registry", expectedRegistry},
		{"toolchain-options", expectedToolchain},
	}
	base := fixture()
	baseDigest := CanonicalInputDigest(base)
	mutations := []struct {
		role string
		edit func(*Input)
	}{
		{"authority-snapshot", func(input *Input) { input.AuthoritySnapshotDigest = "snapshot-mutated" }},
		{"policy", func(input *Input) { input.PolicyDigest = "policy-mutated" }},
		{"registry", func(input *Input) { input.RegistryDigest = "registry-mutated" }},
		{"toolchain-options", func(input *Input) { input.ToolchainOptionsDigest = "toolchain-mutated" }},
	}
	for _, mutation := range mutations {
		input := fixture()
		mutation.edit(&input)
		if CanonicalInputDigest(input) == baseDigest {
			t.Fatalf("%s mutation did not change canonical digest", mutation.role)
		}
		for _, binding := range bindings {
			got := bindingField(input, binding.role)
			if binding.role == mutation.role && got == binding.want {
				t.Fatalf("%s mutation did not change its binding", mutation.role)
			}
			if binding.role != mutation.role && got != binding.want {
				t.Fatalf("%s mutation changed %s", mutation.role, binding.role)
			}
		}
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
		Schema:                  SchemaVersion,
		RequestedK:              21,
		MinimumIndependent:      2,
		AuthoritySnapshotDigest: expectedSnapshot,
		PolicyDigest:            expectedPolicy,
		RegistryDigest:          expectedRegistry,
		ToolchainOptionsDigest:  expectedToolchain,
		PressureRecords: []PressureRecord{
			{"pressure-z", "category-z", "group-z", "rule-1"},
			{"pressure-a", "category-a", "group-a", "rule-1"},
			{"pressure-b", "category-b", "group-b", "rule-1"},
			{"pressure-aa", "category-a", "group-a", "rule-1"},
		},
		RequiredPressureIDs: []string{"pressure-z", "pressure-a", "pressure-b", "pressure-aa"},
	}
	return input
}

func bindingField(input Input, role string) string {
	switch role {
	case "authority-snapshot":
		return input.AuthoritySnapshotDigest
	case "policy":
		return input.PolicyDigest
	case "registry":
		return input.RegistryDigest
	case "toolchain-options":
		return input.ToolchainOptionsDigest
	default:
		return ""
	}
}

func schemaJSON(suffix string) string {
	return `{"schema":"` + SchemaVersion + `"` + suffix + `}`
}
