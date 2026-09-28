package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestSelfImprovementProvenanceProjectionBindsEvidencePrefixBoundary(t *testing.T) {
	unknownChain := provenance.BuildSelfImprovementProvenanceChainPart01(
		"decl", "source", "ir", "graph", "", "reverse",
	)
	unknown, err := ProjectSelfImprovementProvenanceChainPart01(
		"file:///workspace/example.gooo", 4, unknownChain,
	)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if unknown.MissingStageIndex != 4 || unknown.EvidencePrefixDigest == "" {
		t.Fatalf("unknown projection boundary = %+v", unknown)
	}
	tampered := unknown
	tampered.Version++
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered projection unexpectedly validated")
	}

	completeChain := provenance.BuildSelfImprovementProvenanceChainPart01(
		"decl", "source", "ir", "graph", "generated", "reverse",
	)
	complete, err := ProjectSelfImprovementProvenanceChainPart01(
		"file:///workspace/example.gooo", 4, completeChain,
	)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if complete.MissingStageIndex != -1 || complete.EvidencePrefixDigest == unknown.EvidencePrefixDigest {
		t.Fatalf("complete projection boundary = %+v", complete)
	}
}
