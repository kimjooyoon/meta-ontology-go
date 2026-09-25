package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestProjectSelfImprovementProvenanceChainPart01ExposesEvidenceBoundary(t *testing.T) {
	chain := provenance.BuildSelfImprovementProvenanceChainPart01("decl", "source", "ir", "graph", "generated", "reverse")
	projection, err := ProjectSelfImprovementProvenanceChainPart01("file:///workspace/example.gooo", 4, chain)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if projection.Status != provenance.SelfImprovementProvenanceChainBoundPart01 ||
		projection.CausalReason != provenance.SelfImprovementProvenanceChainCompleteReasonPart01 {
		t.Fatalf("projection = %+v", projection)
	}
	if projection.BoundStages != 6 || projection.TotalStages != 6 || projection.NextRequiredStage != "" {
		t.Fatalf("projection stage summary = %+v", projection)
	}
	if projection.AdoptionAuthorized || !projection.NonAuthorizing || projection.Digest == "" {
		t.Fatalf("projection must remain non-authorizing and digestable: %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestProjectSelfImprovementProvenanceChainPart01PreservesUnknown(t *testing.T) {
	chain := provenance.BuildSelfImprovementProvenanceChainPart01("decl", "source", "ir", "graph", "generated", "")
	projection, err := ProjectSelfImprovementProvenanceChainPart01("file:///workspace/example.gooo", 5, chain)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if projection.Status != provenance.SelfImprovementProvenanceChainUnknownPart01 ||
		projection.CausalReason != "MISSING_REVERSE_OBSERVATION_DIGEST" {
		t.Fatalf("projection = %+v", projection)
	}
	if projection.BoundStages != 5 || projection.TotalStages != 6 ||
		projection.NextRequiredStage != "reverse-observation" {
		t.Fatalf("projection stage summary = %+v", projection)
	}
	if err := projection.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}
