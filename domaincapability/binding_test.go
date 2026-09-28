package domaincapability

import "testing"

func TestSourceCapabilityIDsRequireCompleteBinding(t *testing.T) {
	ids := SourceCapabilityIDs([]CapabilityBinding{
		{CapabilityID: "z", DeclarationID: "decl-z", SourceDigest: "sha256:z", EvidenceDigest: "sha256:e1"},
		{CapabilityID: "a", DeclarationID: "decl-a", SourceDigest: "sha256:a", EvidenceDigest: "sha256:e2"},
		{CapabilityID: "missing-evidence", DeclarationID: "decl", SourceDigest: "sha256:s"},
		{CapabilityID: "a", DeclarationID: "decl-a", SourceDigest: "sha256:a", EvidenceDigest: "sha256:e2"},
	})
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "z" {
		t.Fatalf("ids = %#v, want [a z]", ids)
	}
}