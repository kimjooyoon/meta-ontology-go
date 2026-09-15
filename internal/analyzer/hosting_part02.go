package analyzer

import (
	"strings"
)

// ContractFor returns the declared contract for a host stage. Future stages
// are returned as deferred contracts, never as successful implementations.
func ContractFor(stage HostStage) HostingContract {
	switch stage {
	case StageGoHosted:
		return HostingContract{
			Stage:           StageGoHosted,
			Status:          ContractImplemented,
			SourceAuthority: ".gooo",
			Producer:        "go",
			Requirements: []ContractRequirement{
				RequirementStableIdentity,
				RequirementDeltaEvidence,
				RequirementSourceSpans,
			},
		}
	case StageGoooHosted:
		return HostingContract{
			Stage:           StageGoooHosted,
			Status:          ContractDeferred,
			SourceAuthority: ".gooo",
			Producer:        "gooo",
			Requirements: []ContractRequirement{
				RequirementStableIdentity,
				RequirementDeltaEvidence,
				RequirementSourceSpans,
				RequirementHostComparison,
				RequirementIndependentGate,
			},
		}
	default:
		return HostingContract{Stage: stage, Status: ContractDeferred}
	}
}

// Valid reports whether the contract has a known stage and complete metadata.
func (c HostingContract) Valid() bool {
	if !c.Stage.Valid() || !validContractStatus(c.Status) || strings.TrimSpace(c.SourceAuthority) == "" || strings.TrimSpace(c.Producer) == "" {
		return false
	}
	return len(c.Requirements) > 0
}

// PromotionReady reports whether this contract can be treated as implemented.
func (c HostingContract) PromotionReady() bool {
	return c.Valid() && c.Status == ContractImplemented
}
func (s HostStage) Valid() bool {
	return s == StageGoHosted || s == StageGoooHosted
}
func validContractStatus(status ContractStatus) bool {
	return status == ContractImplemented || status == ContractDeferred
}

// EvidenceKind identifies which analyzer view a record preserves.
type EvidenceKind string

const (
	EvidenceKindFact           EvidenceKind = "fact"
	EvidenceKindCandidate      EvidenceKind = "candidate"
	EvidenceKindImplementation EvidenceKind = "implementation"
)

// EvidenceStatus identifies the confidence state of one evidence record.
type EvidenceStatus string
