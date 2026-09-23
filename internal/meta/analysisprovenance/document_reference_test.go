package analysisprovenance

import "testing"

func TestDocumentDigestWithSymbolMapAndReferencesBindsReferenceMap(t *testing.T) {
	first := DocumentDigestWithSymbolMapAndReferences("source", "semantic", "profile", "toolchain", "contract", "symbols", "references-a")
	second := DocumentDigestWithSymbolMapAndReferences("source", "semantic", "profile", "toolchain", "contract", "symbols", "references-b")
	if first == second {
		t.Fatalf("reference map did not change document digest: %q", first)
	}
	if DocumentDigestWithSymbolMap("source", "semantic", "profile", "toolchain", "contract", "symbols") != DocumentDigestWithSymbolMapAndReferences("source", "semantic", "profile", "toolchain", "contract", "symbols", "") {
		t.Fatal("legacy symbol-map digest changed unexpectedly")
	}
}
