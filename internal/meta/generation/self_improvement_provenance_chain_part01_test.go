package generation

import "testing"

func TestBuildSelfImprovementProvenanceChainPart01BindsOrderedEvidence(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01("decl", "src", "ir", "graph", "generated", "reverse")
	if chain.Status != SelfImprovementProvenanceChainBoundPart01 {
		t.Fatalf("status = %q, want %q", chain.Status, SelfImprovementProvenanceChainBoundPart01)
	}
	if chain.CausalReason != SelfImprovementProvenanceChainCompleteReasonPart01 {
		t.Fatalf("reason = %q, want %q", chain.CausalReason, SelfImprovementProvenanceChainCompleteReasonPart01)
	}
	if chain.AdoptionAuthorized || !chain.NonAuthorizing {
		t.Fatalf("chain must remain non-authorizing: %+v", chain)
	}
	if err := chain.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got := BuildSelfImprovementProvenanceChainPart01("decl", "src", "ir", "graph", "generated", "reverse").ChainDigest; got != chain.ChainDigest {
		t.Fatalf("chain digest is not deterministic: %q != %q", got, chain.ChainDigest)
	}
}

func TestBuildSelfImprovementProvenanceChainPart01PreservesUnknownCause(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01("decl", "src", "ir", "graph", "generated", "")
	if chain.Status != SelfImprovementProvenanceChainUnknownPart01 {
		t.Fatalf("status = %q, want %q", chain.Status, SelfImprovementProvenanceChainUnknownPart01)
	}
	if chain.CausalReason != "MISSING_REVERSE_OBSERVATION_DIGEST" {
		t.Fatalf("reason = %q", chain.CausalReason)
	}
	if err := chain.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSelfImprovementProvenanceChainPart01RejectsTampering(t *testing.T) {
	chain := BuildSelfImprovementProvenanceChainPart01("decl", "src", "ir", "graph", "generated", "reverse")
	chain.GeneratedDigest = "changed"
	if err := chain.Validate(); err == nil {
		t.Fatal("Validate() unexpectedly accepted a tampered chain")
	}
}
