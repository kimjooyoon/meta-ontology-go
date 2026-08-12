package bidir

import (
	"fmt"
	"sort"
	"strings"
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

// Validate checks phase identity and evidence uniqueness.
func (c HostingContract) Validate() error {
	if c.Phase != HostPhaseGoHosted && c.Phase != HostPhaseGoooHosted {
		return fmt.Errorf("unknown host phase %d", c.Phase)
	}
	if strings.TrimSpace(c.HostLanguage) == "" {
		return fmt.Errorf("host language is required")
	}
	if strings.TrimSpace(c.AuthoritativeView) == "" {
		return fmt.Errorf("authoritative view is required")
	}
	seen := make(map[string]struct{}, len(c.Evidence))
	for _, evidence := range c.Evidence {
		check := strings.TrimSpace(evidence.Check)
		if check == "" {
			return fmt.Errorf("evidence check is required")
		}
		if evidence.State < EvidencePlanned || evidence.State > EvidenceVerified {
			return fmt.Errorf("evidence %q has unknown state %d", check, evidence.State)
		}
		if _, exists := seen[check]; exists {
			return fmt.Errorf("duplicate evidence check %q", check)
		}
		seen[check] = struct{}{}
	}
	return nil
}

// Verified reports whether every declared check has verified evidence.
func (c HostingContract) Verified() bool {
	if c.Validate() != nil || len(c.Evidence) == 0 {
		return false
	}
	for _, evidence := range c.Evidence {
		if evidence.State != EvidenceVerified {
			return false
		}
	}
	return true
}

// UnverifiedChecks returns planned or observed checks in stable order.
func (c HostingContract) UnverifiedChecks() []string {
	var checks []string
	for _, evidence := range c.Evidence {
		if evidence.State != EvidenceVerified {
			checks = append(checks, strings.TrimSpace(evidence.Check))
		}
	}
	sort.Strings(checks)
	return checks
}

// HostingComparison describes the evidence gap between two phases.
type HostingComparison struct {
	From          HostPhase
	To            HostPhase
	HostChanged   bool
	AddedChecks   []string
	NewlyVerified []string
	Remaining     []string
}

// CompareHostingContracts compares a current contract with a target contract.
func CompareHostingContracts(current, target HostingContract) (HostingComparison, error) {
	if err := current.Validate(); err != nil {
		return HostingComparison{}, fmt.Errorf("current contract: %w", err)
	}
	if err := target.Validate(); err != nil {
		return HostingComparison{}, fmt.Errorf("target contract: %w", err)
	}
	comparison := HostingComparison{From: current.Phase, To: target.Phase, HostChanged: current.HostLanguage != target.HostLanguage}
	currentStates := evidenceStates(current.Evidence)
	for _, evidence := range target.Evidence {
		check := strings.TrimSpace(evidence.Check)
		state, exists := currentStates[check]
		if !exists {
			comparison.AddedChecks = append(comparison.AddedChecks, check)
		}
		if evidence.State == EvidenceVerified && (!exists || state != EvidenceVerified) {
			comparison.NewlyVerified = append(comparison.NewlyVerified, check)
		}
		if evidence.State != EvidenceVerified {
			comparison.Remaining = append(comparison.Remaining, check)
		}
	}
	sort.Strings(comparison.AddedChecks)
	sort.Strings(comparison.NewlyVerified)
	sort.Strings(comparison.Remaining)
	return comparison, nil
}

func evidenceStates(evidence []ContractEvidence) map[string]EvidenceState {
	states := make(map[string]EvidenceState, len(evidence))
	for _, item := range evidence {
		states[strings.TrimSpace(item.Check)] = item.State
	}
	return states
}

// InitialGoHostedContract records the currently executable host boundary.
func InitialGoHostedContract() HostingContract {
	return HostingContract{
		Phase:             HostPhaseGoHosted,
		HostLanguage:      "go",
		AuthoritativeView: ".gooo DSL",
		Evidence: []ContractEvidence{
			{Check: "generic-get-put", State: EvidenceVerified, Detail: "parser-neutral lens tests pass"},
			{Check: "delta-reconcile-locality", State: EvidenceVerified, Detail: "delta and locality tests pass"},
			{Check: "candidate-separation", State: EvidenceVerified, Detail: "candidate facts remain non-deterministic"},
			{Check: "semantic-ir-lowering", State: EvidenceObserved, Detail: "implementation exists; prerequisite package integration is pending"},
			{Check: "go-projection-reconcile", State: EvidencePlanned, Detail: "requires integrated analyzer and generator lanes"},
			{Check: "gooo-hosted-self-hosting", State: EvidencePlanned, Detail: "future phase; not implemented"},
		},
	}
}

// PlannedGoooHostedContract describes the future phase without claiming it is
// complete. Existing verified checks are carried forward; host transition and
// self-hosting remain explicit evidence gaps.
func PlannedGoooHostedContract() HostingContract {
	contract := InitialGoHostedContract()
	contract.Phase = HostPhaseGoooHosted
	contract.HostLanguage = "gooo"
	return contract
}
