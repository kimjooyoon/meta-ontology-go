package bidir

import (
	"fmt"
)

// HostPhase names the implementation host for a compiler milestone.
type HostPhase uint8

const (
	HostPhaseGoHosted HostPhase = iota + 1
	HostPhaseGoooHosted

	// GoHostedPhase and GoooHostedPhase are concise aliases for callers.
	GoHostedPhase   = HostPhaseGoHosted
	GoooHostedPhase = HostPhaseGoooHosted
)

func (p HostPhase) String() string {
	switch p {
	case HostPhaseGoHosted:
		return "go-hosted"
	case HostPhaseGoooHosted:
		return "gooo-hosted"
	default:
		return fmt.Sprintf("host-phase(%d)", p)
	}
}

// EvidenceState distinguishes a planned check from verified evidence.
type EvidenceState uint8

const (
	EvidencePlanned EvidenceState = iota + 1
	EvidenceObserved
	EvidenceVerified
)

func (s EvidenceState) String() string {
	switch s {
	case EvidencePlanned:
		return "planned"
	case EvidenceObserved:
		return "observed"
	case EvidenceVerified:
		return "verified"
	default:
		return fmt.Sprintf("evidence-state(%d)", s)
	}
}

// ContractEvidence is one check supporting a hosting contract.
type ContractEvidence struct {
	Check  string
	State  EvidenceState
	Detail string
}

// HostingContract makes host-phase claims explicit and reviewable.
type HostingContract struct {
	Phase             HostPhase
	HostLanguage      string
	AuthoritativeView string
	Evidence          []ContractEvidence
}
