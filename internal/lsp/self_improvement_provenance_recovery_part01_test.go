package lsp

import (
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/provenance"
)

func TestProjectSelfImprovementProvenanceRecoveryBindsRequiredEvidence(t *testing.T) {
	unknownChain := provenance.BuildSelfImprovementProvenanceChainPart01(
		"decl", "source", "ir", "graph", "generated", "",
	)
	unknown, err := ProjectSelfImprovementProvenanceChainPart01(
		"file:///workspace/example.gooo", 4, unknownChain,
	)
	if err != nil {
		t.Fatalf("ProjectSelfImprovementProvenanceChainPart01() error = %v", err)
	}
	if unknown.Recovery.Stage != "reverse-observation" ||
		unknown.Recovery.StageIndex != 5 ||
		unknown.Recovery.RequiredEvidence != "REVERSE_OBSERVATION_DIGEST" ||
		unknown.Recovery.Action != "BIND_REVERSE_OBSERVATION_DIGEST" ||
		unknown.Recovery.Digest == "" {
		t.Fatalf("unknown recovery = %+v", unknown.Recovery)
	}
	if err := unknown.Validate(); err != nil {
		t.Fatalf("unknown projection should validate: %v", err)
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
	if complete.Recovery.StageIndex != -1 || complete.Recovery.Action != "NO_ACTION_REQUIRED" ||
		!complete.Recovery.NonAuthorizing {
		t.Fatalf("complete recovery = %+v", complete.Recovery)
	}
	if err := complete.Validate(); err != nil {
		t.Fatalf("complete projection should validate: %v", err)
	}

	tampered := unknown
	tampered.Recovery.Action = "AUTHORIZE_ADOPTION"
	if err := tampered.Validate(); err == nil {
		t.Fatal("tampered recovery unexpectedly validated")
	}
}
