package pressurecoverage

import (
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
