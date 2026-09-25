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

func TestExecutionPlanBindingPreservesTypedPlanIdentity(t *testing.T) {
	digest := "sha256:" + strings.Repeat("e", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	activities := []string{"runtimebinding.ProposeCandidate", "runtimebinding.RecordIndependentReview", "runtimebinding.CommitCandidate"}
	binding := BindExecutionPlanToProvenanceWithTypedPlanPart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{},
		ExecutionPlanLifecycleResumed,
		digest,
		activities,
		2,
		chain,
	)
	if binding.Plan.TypedPlanDigest != digest || binding.Plan.RuntimeBindingCount != 2 || len(binding.Plan.ActivityOrder) != len(activities) {
		t.Fatalf("typed plan metadata = %#v", binding.Plan)
	}
	if binding.Status != ExecutionPlanBindingBound {
		t.Fatalf("typed plan binding = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("typed plan binding validation failed: %v", err)
	}
	binding.Plan.ActivityOrder[1] = "tampered"
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered typed activity order was accepted")
	}
}

func TestExecutionPlanBindingPreservesTypedPlanEdgeIdentity(t *testing.T) {
	digest := "sha256:" + strings.Repeat("e", 64)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	activities := []string{"runtimebinding.ProposeCandidate", "runtimebinding.RecordIndependentReview", "runtimebinding.CommitCandidate"}
	edges := []string{
		"runtimebinding.ProposeCandidate:result->runtimebinding.RecordIndependentReview:input",
		"runtimebinding.RecordIndependentReview:result->runtimebinding.CommitCandidate:input",
	}
	binding := BindExecutionPlanToProvenanceWithTypedPlanEdgesPart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{},
		ExecutionPlanLifecycleResumed,
		digest,
		activities,
		edges,
		2,
		chain,
	)
	if len(binding.Plan.BindingEdgeOrder) != len(edges) {
		t.Fatalf("binding edge order = %#v", binding.Plan.BindingEdgeOrder)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("typed plan edge binding validation failed: %v", err)
	}
	binding.Plan.BindingEdgeOrder[0] = "tampered"
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered typed plan edge order was accepted")
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

func TestExecutionPlanBindingPreservesWorkloadIdentityObservation(t *testing.T) {
	digest := "sha256:" + strings.Repeat("a", 64)
	identity := WorkloadIdentityProvenanceBinding{
		Schema:                 WorkloadIdentityProvenanceBindingSchema,
		ProvenanceChainDigest:  digest,
		IdentityEvidenceDigest: digest,
		Status:                 WorkloadIdentityProvenanceBindingObserved,
		Reason:                 "PROVENANCE_AND_IDENTITY_EVIDENCE_BOUND",
		NonAuthorizing:         true,
	}
	identity.BindingDigest = workloadIdentityProvenanceBindingDigest(identity)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanToProvenanceWithWorkloadIdentityPart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{AllowedHosts: []string{"api.example.com"}},
		ExecutionPlanLifecycleResumed,
		identity,
		chain,
	)
	if binding.Status != ExecutionPlanBindingBound ||
		binding.CausalReason != "EXECUTION_PLAN_PROVENANCE_AND_IDENTITY_BOUND" ||
		binding.Plan.WorkloadIdentityBindingDigest != identity.BindingDigest {
		t.Fatalf("workload identity execution binding = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("workload identity execution binding validation failed: %v", err)
	}
	binding.Plan.WorkloadIdentityBindingDigest = "sha256:" + strings.Repeat("f", 64)
	if err := binding.Validate(); err == nil {
		t.Fatal("tampered workload identity binding digest was accepted")
	}
}

func TestExecutionPlanBindingKeepsUnknownWorkloadIdentityUnknown(t *testing.T) {
	digest := "sha256:" + strings.Repeat("b", 64)
	identity := WorkloadIdentityProvenanceBinding{
		Schema:                 WorkloadIdentityProvenanceBindingSchema,
		ProvenanceChainDigest:  digest,
		IdentityEvidenceDigest: digest,
		Status:                 WorkloadIdentityProvenanceBindingUnknown,
		Reason:                 "IDENTITY_EVIDENCE_NOT_OBSERVED",
		NonAuthorizing:         true,
	}
	identity.BindingDigest = workloadIdentityProvenanceBindingDigest(identity)
	chain := BuildSelfImprovementProvenanceChainPart01(digest, digest, digest, digest, digest, digest)
	binding := BindExecutionPlanToProvenanceWithWorkloadIdentityPart01(
		"compile-gooo",
		digest,
		"model/gooo-planner-v1",
		GatewayPolicy{},
		ExecutionPlanLifecyclePlanned,
		identity,
		chain,
	)
	if binding.Status != ExecutionPlanBindingUnknown ||
		binding.CausalReason != "EXECUTION_PLAN_WORKLOAD_IDENTITY_UNKNOWN" ||
		binding.Plan.WorkloadIdentityBindingDigest != identity.BindingDigest {
		t.Fatalf("unknown workload identity execution binding = %#v", binding)
	}
	if err := binding.Validate(); err != nil {
		t.Fatalf("unknown workload identity execution binding validation failed: %v", err)
	}
}
