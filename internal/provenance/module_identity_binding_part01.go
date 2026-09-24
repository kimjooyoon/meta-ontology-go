package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

const ModuleIdentityProvenanceBindingSchema = "gooo/module-identity-provenance-binding/v1"

type ModuleIdentityObservationStatus string

const (
	ModuleIdentityObserved ModuleIdentityObservationStatus = "OBSERVED"
	ModuleIdentityUnknown  ModuleIdentityObservationStatus = "UNKNOWN"
	ModuleIdentityRefuted  ModuleIdentityObservationStatus = "REFUTED"
)

type ModuleReleaseIdentity struct {
	ModulePath string `json:"module_path"`
	Release    string `json:"release"`
	Digest     string `json:"digest"`
}

func (i ModuleReleaseIdentity) Validate() error {
	if strings.TrimSpace(i.ModulePath) == "" {
		return fmt.Errorf("module path must be non-empty")
	}
	if strings.TrimSpace(i.Release) == "" {
		return fmt.Errorf("module release must be non-empty")
	}
	if strings.TrimSpace(i.Digest) == "" {
		return fmt.Errorf("module digest must be non-empty")
	}
	return nil
}

type ModuleIdentityProvenanceBinding struct {
	SchemaVersion      string                         `json:"schema_version"`
	Identity           ModuleReleaseIdentity          `json:"identity"`
	SourceDigest       string                         `json:"source_digest,omitempty"`
	SemanticDigest     string                         `json:"semantic_digest,omitempty"`
	PlanDigest         string                         `json:"plan_digest,omitempty"`
	BindingDigest      string                         `json:"binding_digest,omitempty"`
	Status             ModuleIdentityObservationStatus `json:"status"`
	MissingStageIndex  int                            `json:"missing_stage_index"`
	MissingStage       string                         `json:"missing_stage,omitempty"`
	UnknownReason      string                         `json:"unknown_reason,omitempty"`
}

func NewModuleIdentityProvenanceBindingPart01() ModuleIdentityProvenanceBinding {
	return ModuleIdentityProvenanceBinding{
		SchemaVersion:     ModuleIdentityProvenanceBindingSchema,
		MissingStageIndex: -1,
	}
}

func (b *ModuleIdentityProvenanceBinding) preserveUnknown(stageIndex int, stage, reason string) {
	if b.MissingStageIndex >= 0 {
		return
	}
	b.MissingStageIndex = stageIndex
	b.MissingStage = stage
	b.UnknownReason = reason
	b.Status = ModuleIdentityUnknown
}

func BindModuleIdentityToPlanPart01(
	observed, expected ModuleReleaseIdentity,
	sourceDigest, semanticDigest, planDigest string,
) ModuleIdentityProvenanceBinding {
	binding := NewModuleIdentityProvenanceBindingPart01()
	binding.Identity = observed
	binding.SourceDigest = sourceDigest
	binding.SemanticDigest = semanticDigest
	binding.PlanDigest = planDigest

	if err := observed.Validate(); err != nil {
		binding.preserveUnknown(0, "module-identity", err.Error())
		return binding
	}
	if err := expected.Validate(); err != nil {
		binding.preserveUnknown(0, "expected-module-identity", err.Error())
		return binding
	}
	if observed != expected {
		binding.Status = ModuleIdentityRefuted
		binding.UnknownReason = "observed module release identity differs from expected identity"
		return binding
	}
	switch {
	case strings.TrimSpace(sourceDigest) == "":
		binding.preserveUnknown(1, "source", "source digest is missing")
	case strings.TrimSpace(semanticDigest) == "":
		binding.preserveUnknown(2, "semantic", "semantic digest is missing")
	case strings.TrimSpace(planDigest) == "":
		binding.preserveUnknown(3, "plan", "execution plan digest is missing")
	default:
		binding.Status = ModuleIdentityObserved
		binding.BindingDigest = binding.computeDigest()
	}
	return binding
}

func (b ModuleIdentityProvenanceBinding) Validate() error {
	if b.SchemaVersion != ModuleIdentityProvenanceBindingSchema {
		return fmt.Errorf("unexpected module identity provenance schema %q", b.SchemaVersion)
	}
	if err := b.Identity.Validate(); err != nil {
		return err
	}
	switch b.Status {
	case ModuleIdentityObserved:
		if strings.TrimSpace(b.SourceDigest) == "" ||
			strings.TrimSpace(b.SemanticDigest) == "" ||
			strings.TrimSpace(b.PlanDigest) == "" {
			return fmt.Errorf("OBSERVED module identity binding requires source, semantic, and plan digests")
		}
		if b.BindingDigest != b.computeDigest() {
			return fmt.Errorf("module identity binding digest mismatch")
		}
	case ModuleIdentityUnknown:
		if b.MissingStageIndex < 0 ||
			strings.TrimSpace(b.MissingStage) == "" ||
			strings.TrimSpace(b.UnknownReason) == "" {
			return fmt.Errorf("UNKNOWN module identity binding requires first missing stage coordinates")
		}
	case ModuleIdentityRefuted:
		if strings.TrimSpace(b.UnknownReason) == "" {
			return fmt.Errorf("REFUTED module identity binding requires a reason")
		}
	default:
		return fmt.Errorf("invalid module identity observation status %q", b.Status)
	}
	return nil
}

func (b ModuleIdentityProvenanceBinding) computeDigest() string {
	payload := struct {
		SchemaVersion  string                `json:"schema_version"`
		Identity       ModuleReleaseIdentity `json:"identity"`
		SourceDigest   string                `json:"source_digest"`
		SemanticDigest string                `json:"semantic_digest"`
		PlanDigest     string                `json:"plan_digest"`
	}{
		SchemaVersion:  b.SchemaVersion,
		Identity:       b.Identity,
		SourceDigest:   b.SourceDigest,
		SemanticDigest: b.SemanticDigest,
		PlanDigest:     b.PlanDigest,
	}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}
