package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestProjectSelfImprovementProvenanceChainPart01ExposesEvidenceBoundary(t *testing.T) {
	chain := generation.BuildSelfImprovementProvenanceChainPart01("decl", "source", "ir", "graph", "generated", "reverse")
	projection, err := ProjectSelfImprovementProvenanceChainPart01("file:///workspace/example.gooo", 4, chain)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if projection.Status != generation.SelfImprovementProvenanceChainBoundPart01 || projection.CausalReason != generation.SelfImprovementProvenanceChainCompleteReasonPart01 {
		t.Fatalf("projection = %+v", projection)
	}
	if projection.AdoptionAuthorized || !projection.NonAuthorizing || projection.Digest == "" {
		t.Fatalf("projection must remain non-authorizing and digestable: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectSelfImprovementProvenanceChainPart01PreservesUnknown(t *testing.T) {
	chain := generation.BuildSelfImprovementProvenanceChainPart01("decl", "source", "ir", "graph", "generated", "")
	projection, err := ProjectSelfImprovementProvenanceChainPart01("file:///workspace/example.gooo", 5, chain)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if projection.Status != generation.SelfImprovementProvenanceChainUnknownPart01 || projection.CausalReason != "MISSING_REVERSE_OBSERVATION_DIGEST" {
		t.Fatalf("projection = %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}