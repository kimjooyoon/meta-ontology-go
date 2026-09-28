package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestSemanticAdoptionObservationIdentityBindsLoweredIR(t *testing.T) {
	digest := func(value string) string { return cache.HashBytes([]byte(value)).String() }
	provenance := SemanticAdoptionProvenance{
		SourceDigest: digest("source"), ProfileDigest: digest("profile"),
		ToolchainDigest: digest("toolchain"), ContractDigest: digest("contract"),
	}
	identity := provenance.ObservationIdentity(digest("semantic"))
	if err := identity.Validate(); err != nil {
		t.Fatal(err)
	}
	if identity.SubjectDigest != provenance.SourceDigest || identity.SemanticDigest != digest("semantic") {
		t.Fatalf("observation identity lost source or semantic binding: %#v", identity)
	}
}
