package analysisprovenance

import "testing"

func TestDocumentDigestBindsSemanticLowering(t *testing.T) {
	base := DocumentDigest("source", "semantic", "profile", "toolchain", "contract")
	if base == "" || base != DocumentDigest("source", "semantic", "profile", "toolchain", "contract") {
		t.Fatalf("document digest is not deterministic: %q", base)
	}
	if base == DocumentDigest("source", "changed-semantic", "profile", "toolchain", "contract") {
		t.Fatal("document digest ignored semantic lowering")
	}
}
