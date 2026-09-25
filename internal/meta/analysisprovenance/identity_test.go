package analysisprovenance

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestObservationIdentityBindsSemanticAndAnalysisInputs(t *testing.T) {
	digest := func(value string) string { return cache.HashBytes([]byte(value)).String() }
	identity := NewObservationIdentity(digest("source"), digest("semantic"), digest("profile"), digest("toolchain"), digest("contract"))
	if err := identity.Validate(); err != nil {
		t.Fatal(err)
	}
	changed := NewObservationIdentity(identity.SubjectDigest, digest("different-semantic"), identity.ProfileDigest, identity.ToolchainDigest, identity.ContractDigest)
	if changed.ProvenanceDigest == identity.ProvenanceDigest {
		t.Fatal("semantic digest change did not change observation identity")
	}
}

func TestObservationIdentityRejectsTampering(t *testing.T) {
	digest := func(value string) string { return cache.HashBytes([]byte(value)).String() }
	identity := NewObservationIdentity(digest("source"), digest("semantic"), digest("profile"), digest("toolchain"), digest("contract"))
	identity.SemanticDigest = digest("tampered")
	if identity.Valid() {
		t.Fatal("tampered semantic digest was accepted")
	}
}
