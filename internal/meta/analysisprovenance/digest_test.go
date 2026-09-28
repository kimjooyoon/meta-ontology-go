package analysisprovenance

import "testing"

func TestDigestBindsOrderedProvenanceComponents(t *testing.T) {
	base := Digest("source", "profile", "toolchain", "contract")
	if base == "" {
		t.Fatal("digest is empty")
	}
	if base == Digest("profile", "source", "toolchain", "contract") {
		t.Fatal("digest ignored component order")
	}
	if base == Digest("source", "profile", "toolchain", "other-contract") {
		t.Fatal("digest ignored contract identity")
	}
}
