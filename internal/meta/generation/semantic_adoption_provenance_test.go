package generation

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/cache"
)

func TestSemanticAdoptionProvenanceRequiresFourKnownDigests(t *testing.T) {
	valid := &SemanticAdoptionProvenance{
		SourceDigest: cache.HashBytes([]byte("source")).String(), ProfileDigest: cache.HashBytes([]byte("profile")).String(),
		ToolchainDigest: cache.HashBytes([]byte("toolchain")).String(), ContractDigest: cache.HashBytes([]byte("contract")).String(),
	}
	if !validSemanticAdoptionProvenance(valid) {
		t.Fatal("known adoption provenance was rejected")
	}
	invalid := *valid
	invalid.ToolchainDigest = "unknown"
	if validSemanticAdoptionProvenance(&invalid) {
		t.Fatal("unknown toolchain provenance was accepted")
	}
}
