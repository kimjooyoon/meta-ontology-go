package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

type CapabilityEffect string

const (
	EffectReadInput            CapabilityEffect = "READ_INPUT"
	EffectNetworkReadPinned    CapabilityEffect = "NETWORK_READ_PINNED"
	EffectGenerateCallerOutput CapabilityEffect = "GENERATE_CALLER_OUTPUT"
	EffectRepositoryWrite      CapabilityEffect = "REPOSITORY_WRITE"
	EffectRemoteMutation       CapabilityEffect = "REMOTE_MUTATION"
	EffectDestructiveDelete    CapabilityEffect = "DESTRUCTIVE_DELETE"
)

type CapabilityEffectBindingStatus string

const (
	CapabilityEffectObserved CapabilityEffectBindingStatus = "OBSERVED"
	CapabilityEffectUnknown  CapabilityEffectBindingStatus = "UNKNOWN"
	CapabilityEffectRefuted  CapabilityEffectBindingStatus = "REFUTED"
)

type CapabilityEffectEvidence struct {
	GeneratedArtifactDigest  string
	ReverseObservationDigest string
	MetricEvidenceDigest     string
}

type CapabilityEffectBindingRequest struct {
	SourceDeclarationPath   string
	SourceDeclarationDigest string
	SemanticIRDigest       string
	ExecutionPlanDigest    string
	WorkloadIdentityDigest string
	ScopeDigest            string
	Effects                []CapabilityEffect
	AttenuationEdgeDigests []string
	StageIndex             int
	ParentStageIndex       int
	ParentEffects          []CapabilityEffect
	ParentEffectDigest     string
	Evidence               CapabilityEffectEvidence
}

type CapabilityEffectProvenanceBinding struct {
	SourceDeclarationPath   string
	SourceDeclarationDigest string
	SemanticIRDigest       string
	ExecutionPlanDigest    string
	WorkloadIdentityDigest string
	ScopeDigest            string
	Effects                []CapabilityEffect
	AttenuationEdgeDigests []string
	StageIndex             int
	ParentStageIndex       int
	ParentEffectDigest     string
	Evidence               CapabilityEffectEvidence
	Status                 CapabilityEffectBindingStatus
	FirstUnknownStage      string
	Reason                 string
	BindingDigest          string
}

func BindCapabilityEffectProvenancePart01(request CapabilityEffectBindingRequest) CapabilityEffectProvenanceBinding {
	binding := CapabilityEffectProvenanceBinding{
		SourceDeclarationPath:   request.SourceDeclarationPath,
		SourceDeclarationDigest: request.SourceDeclarationDigest,
		SemanticIRDigest:       request.SemanticIRDigest,
		ExecutionPlanDigest:    request.ExecutionPlanDigest,
		WorkloadIdentityDigest: request.WorkloadIdentityDigest,
		ScopeDigest:            request.ScopeDigest,
		Effects:                append([]CapabilityEffect(nil), request.Effects...),
		AttenuationEdgeDigests: append([]string(nil), request.AttenuationEdgeDigests...),
		StageIndex:             request.StageIndex,
		ParentStageIndex:       request.ParentStageIndex,
		ParentEffectDigest:     request.ParentEffectDigest,
		Evidence:               request.Evidence,
	}

	if binding.SourceDeclarationPath == "" || binding.SourceDeclarationDigest == "" {
		return finishUnknown(binding, "source_declaration", "CAPABILITY_EFFECT_SOURCE_DECLARATION_UNKNOWN")
	}
	if binding.SemanticIRDigest == "" {
		return finishUnknown(binding, "semantic_ir", "CAPABILITY_EFFECT_SEMANTIC_IR_UNKNOWN")
	}
	if binding.ExecutionPlanDigest == "" {
		return finishUnknown(binding, "execution_plan", "CAPABILITY_EFFECT_EXECUTION_PLAN_UNKNOWN")
	}
	if binding.WorkloadIdentityDigest == "" {
		return finishUnknown(binding, "workload_identity", "CAPABILITY_EFFECT_WORKLOAD_IDENTITY_UNKNOWN")
	}
	if binding.ScopeDigest == "" {
		return finishUnknown(binding, "scope", "CAPABILITY_EFFECT_SCOPE_UNKNOWN")
	}
	if binding.StageIndex < 0 {
		return finishUnknown(binding, "stage", "CAPABILITY_EFFECT_STAGE_UNKNOWN")
	}
	if reason := validateEffects(binding.Effects); reason != "" {
		return finishRefuted(binding, reason)
	}
	if len(binding.Effects) == 0 {
		return finishUnknown(binding, "effect_declaration", "CAPABILITY_EFFECT_DECLARATION_UNKNOWN")
	}
	if binding.ParentStageIndex >= 0 {
		if binding.StageIndex <= binding.ParentStageIndex {
			return finishRefuted(binding, "CAPABILITY_EFFECT_STAGE_ORDER_REFUTED")
		}
		if binding.ParentEffectDigest == "" {
			return finishUnknown(binding, "parent_effect_observation", "CAPABILITY_EFFECT_PARENT_UNKNOWN")
		}
		if EffectSetDigest(binding.ParentEffects) != binding.ParentEffectDigest {
			return finishRefuted(binding, "CAPABILITY_EFFECT_PARENT_DIGEST_REFUTED")
		}
		if !effectSubset(binding.Effects, binding.ParentEffects) {
			return finishRefuted(binding, "CAPABILITY_EFFECT_WIDENING_REFUTED")
		}
	}
	if binding.Evidence.GeneratedArtifactDigest == "" {
		return finishUnknown(binding, "generation", "CAPABILITY_EFFECT_GENERATION_UNKNOWN")
	}
	if binding.Evidence.ReverseObservationDigest == "" {
		return finishUnknown(binding, "reverse_observation", "CAPABILITY_EFFECT_REVERSE_OBSERVATION_UNKNOWN")
	}
	if binding.Evidence.MetricEvidenceDigest == "" {
		return finishUnknown(binding, "metrics", "CAPABILITY_EFFECT_METRICS_UNKNOWN")
	}
	binding.Status = CapabilityEffectObserved
	binding.Reason = "CAPABILITY_EFFECT_PROVENANCE_OBSERVED"
	binding.BindingDigest = capabilityEffectBindingDigest(binding)
	return binding
}

func EffectSetDigest(effects []CapabilityEffect) string {
	values := make([]string, 0, len(effects))
	for _, effect := range effects {
		values = append(values, string(effect))
	}
	sort.Strings(values)
	return digestStrings(values)
}

func ValidateCapabilityEffectProvenanceBindingPart01(binding CapabilityEffectProvenanceBinding) error {
	if binding.BindingDigest == "" {
		return fmt.Errorf("capability effect binding digest is missing")
	}
	if want := capabilityEffectBindingDigest(binding); want != binding.BindingDigest {
		return fmt.Errorf("capability effect binding digest mismatch")
	}
	switch binding.Status {
	case CapabilityEffectObserved:
		if binding.FirstUnknownStage != "" {
			return fmt.Errorf("observed capability effect binding has unknown stage")
		}
	case CapabilityEffectUnknown:
		if binding.FirstUnknownStage == "" {
			return fmt.Errorf("unknown capability effect binding has no first unknown stage")
		}
	case CapabilityEffectRefuted:
		if binding.Reason == "" {
			return fmt.Errorf("refuted capability effect binding has no reason")
		}
	default:
		return fmt.Errorf("unknown capability effect binding status %q", binding.Status)
	}
	return nil
}

func finishUnknown(binding CapabilityEffectProvenanceBinding, stage, reason string) CapabilityEffectProvenanceBinding {
	binding.Status = CapabilityEffectUnknown
	binding.FirstUnknownStage = stage
	binding.Reason = reason
	binding.BindingDigest = capabilityEffectBindingDigest(binding)
	return binding
}

func finishRefuted(binding CapabilityEffectProvenanceBinding, reason string) CapabilityEffectProvenanceBinding {
	binding.Status = CapabilityEffectRefuted
	binding.Reason = reason
	binding.BindingDigest = capabilityEffectBindingDigest(binding)
	return binding
}

func validateEffects(effects []CapabilityEffect) string {
	seen := map[CapabilityEffect]bool{}
	for _, effect := range effects {
		switch effect {
		case EffectReadInput, EffectNetworkReadPinned, EffectGenerateCallerOutput, EffectRepositoryWrite, EffectRemoteMutation, EffectDestructiveDelete:
		default:
			return "CAPABILITY_EFFECT_KIND_REFUTED"
		}
		if seen[effect] {
			return "CAPABILITY_EFFECT_DUPLICATE_REFUTED"
		}
		seen[effect] = true
	}
	return ""
}

func effectSubset(child, parent []CapabilityEffect) bool {
	allowed := map[CapabilityEffect]bool{}
	for _, effect := range parent {
		allowed[effect] = true
	}
	for _, effect := range child {
		if !allowed[effect] {
			return false
		}
	}
	return true
}

func capabilityEffectBindingDigest(binding CapabilityEffectProvenanceBinding) string {
	effects := make([]string, 0, len(binding.Effects))
	for _, effect := range binding.Effects {
		effects = append(effects, string(effect))
	}
	sort.Strings(effects)
	edges := append([]string(nil), binding.AttenuationEdgeDigests...)
	sort.Strings(edges)
	return digestStrings([]string{
		binding.SourceDeclarationPath,
		binding.SourceDeclarationDigest,
		binding.SemanticIRDigest,
		binding.ExecutionPlanDigest,
		binding.WorkloadIdentityDigest,
		binding.ScopeDigest,
		strings.Join(effects, ","),
		strings.Join(edges, ","),
		fmt.Sprintf("%d", binding.StageIndex),
		fmt.Sprintf("%d", binding.ParentStageIndex),
		binding.ParentEffectDigest,
		binding.Evidence.GeneratedArtifactDigest,
		binding.Evidence.ReverseObservationDigest,
		binding.Evidence.MetricEvidenceDigest,
		string(binding.Status),
		binding.FirstUnknownStage,
		binding.Reason,
	})
}

func digestStrings(values []string) string {
	h := sha256.New()
	for _, value := range values {
		h.Write([]byte(value))
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}
