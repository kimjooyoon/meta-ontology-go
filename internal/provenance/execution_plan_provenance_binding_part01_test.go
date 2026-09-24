package provenance

import (
	"strings"
	"testing"
)

func TestBindExecutionPlanToCompleteProvenance(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanToProvenancePart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{AllowedHosts: []string{"registry.example.com", "api.example.com"}},
		ExecutionPlanLifecycleSuspended,
		chain,
	)
	if binding.Status != ExecutionPlanBindingBound || binding.CausalReason != "EXECUTION_PLAN_PROVENANCE_BOUND" {
		t.Fatalf("binding = %#v", binding)
	}
	if binding.Plan.Lifecycle != ExecutionPlanLifecycleSuspended {
		t.Fatalf("lifecycle = %q", binding.Plan.Lifecycle)
	}
	if binding.AdoptionAuthorized || !binding.NonAuthorizing {
		t.Fatalf("authority flags = %#v", binding)
	}
	if binding.BoundStages != 6 || binding.TotalStages != 6 ||
		binding.MissingStageIndex != -1 || binding.NextRequiredStage != "" ||
		binding.EvidencePrefixDigest == "" {
		t.Fatalf("complete provenance boundary = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("binding validation failed: %v", err)
	}
}

func TestBindExecutionPlanKeepsUnknownProvenanceUnknown(t *testing.T) {
	digest := "sha256:" + strings.Repeat("b", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, "", digest)
	binding := BindExecutionPlanToProvenancePart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{AllowedHosts: []string{"registry.example.com"}},
		ExecutionPlanLifecyclePlanned,
		chain,
	)
	if binding.Status != ExecutionPlanBindingUnknown || binding.CausalReason != "PROVENANCE_CHAIN_UNKNOWN:MISSING_GENERATED_DIGEST" {
		t.Fatalf("binding = %#v", binding)
	}
	if binding.BoundStages != 4 || binding.TotalStages != 6 ||
		binding.MissingStageIndex != 4 || binding.NextRequiredStage != "generated" ||
		binding.EvidencePrefixDigest == "" {
		t.Fatalf("unknown provenance boundary = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("unknown binding validation failed: %v", err)
	}
}

func TestExecutionPlanBindingRejectsTamperedEvidence(t *testing.T) {
	digest := "sha256:" + strings.Repeat("c", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanToProvenancePart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{},
		ExecutionPlanLifecycleResumed,
		chain,
	)
	binding.BindingDigest = "sha256:" + strings.Repeat("f", 64)
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered binding was accepted")
	}
	binding = BindExecutionPlanToProvenancePart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{},
		ExecutionPlanLifecycleResumed,
		chain,
	)
	binding.EvidencePrefixDigest = "sha256:" + strings.Repeat("e", 64)
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered evidence prefix was accepted")
	}
}

func TestExecutionPlanInvalidBoundaryRemainsUnknown(t *testing.T) {
	digest := "sha256:" + strings.Repeat("d", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanToProvenancePart01("", "not-a-digest", "", GatewayPolicy{}, "BROKEN", chain)
	if binding.Status != ExecutionPlanBindingUnknown || binding.Plan.Lifecycle != ExecutionPlanLifecycleUnknown {
		t.Fatalf("binding = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("unknown boundary validation failed: %v", err)
	}
}
