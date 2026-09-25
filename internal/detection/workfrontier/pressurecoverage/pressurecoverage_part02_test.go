package pressurecoverage

import (
	"bytes"
	"testing"
)

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
