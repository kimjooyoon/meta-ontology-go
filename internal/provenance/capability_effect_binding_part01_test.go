package provenance

import "testing"

func completeCapabilityEffectRequest() CapabilityEffectBindingRequest {
	return CapabilityEffectBindingRequest{
		SourceDeclarationPath:   "internal/meta/generation/capability_effect_provenance_contract.gooo",
		SourceDeclarationDigest: "sha256:source",
		SemanticIRDigest:       "sha256:ir",
		ExecutionPlanDigest:    "sha256:plan",
		WorkloadIdentityDigest: "sha256:subject",
		ScopeDigest:            "sha256:scope",
		Effects:                []CapabilityEffect{EffectReadInput},
		AttenuationEdgeDigests: []string{"sha256:edge"},
		StageIndex:             2,
		ParentStageIndex:       -1,
		Evidence: CapabilityEffectEvidence{
			GeneratedArtifactDigest:  "sha256:generation",
			ReverseObservationDigest: "sha256:reverse",
			MetricEvidenceDigest:     "sha256:metrics",
		},
	}
}

func TestCapabilityEffectProvenanceBindingObservesAllStages(t *testing.T) {
	binding := BindCapabilityEffectProvenancePart01(completeCapabilityEffectRequest())
	if binding.Status != CapabilityEffectObserved {
		t.Fatalf("status = %q, want %q", binding.Status, CapabilityEffectObserved)
	}
	if binding.FirstUnknownStage != "" {
		t.Fatalf("first unknown stage = %q", binding.FirstUnknownStage)
	}
	if err := ValidateCapabilityEffectProvenanceBindingPart01(binding); err != nil {
		t.Fatalf("validate observed binding: %v", err)
	}
}

func TestCapabilityEffectProvenanceBindingPreservesFirstMissingReverseStage(t *testing.T) {
	request := completeCapabilityEffectRequest()
	request.Evidence.ReverseObservationDigest = ""
	request.Evidence.MetricEvidenceDigest = ""
	binding := BindCapabilityEffectProvenancePart01(request)
	if binding.Status != CapabilityEffectUnknown {
		t.Fatalf("status = %q, want %q", binding.Status, CapabilityEffectUnknown)
	}
	if binding.FirstUnknownStage != "reverse_observation" {
		t.Fatalf("first unknown stage = %q, want reverse_observation", binding.FirstUnknownStage)
	}
	if err := ValidateCapabilityEffectProvenanceBindingPart01(binding); err != nil {
		t.Fatalf("validate unknown binding: %v", err)
	}
}

func TestCapabilityEffectProvenanceBindingRefutesWidening(t *testing.T) {
	request := completeCapabilityEffectRequest()
	request.ParentStageIndex = 1
	request.ParentEffects = []CapabilityEffect{EffectReadInput}
	request.ParentEffectDigest = EffectSetDigest(request.ParentEffects)
	request.Effects = []CapabilityEffect{EffectRemoteMutation}
	binding := BindCapabilityEffectProvenancePart01(request)
	if binding.Status != CapabilityEffectRefuted {
		t.Fatalf("status = %q, want %q", binding.Status, CapabilityEffectRefuted)
	}
	if binding.Reason != "CAPABILITY_EFFECT_WIDENING_REFUTED" {
		t.Fatalf("reason = %q", binding.Reason)
	}
}

func TestCapabilityEffectProvenanceBindingRejectsTampering(t *testing.T) {
	binding := BindCapabilityEffectProvenancePart01(completeCapabilityEffectRequest())
	binding.ExecutionPlanDigest = "sha256:tampered"
	if err := ValidateCapabilityEffectProvenanceBindingPart01(binding); err == nil {
		t.Fatal("tampered binding unexpectedly validated")
	}
}
