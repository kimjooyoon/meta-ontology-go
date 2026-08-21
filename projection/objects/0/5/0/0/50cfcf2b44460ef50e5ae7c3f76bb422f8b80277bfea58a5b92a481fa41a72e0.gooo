package bidir

import (
	"fmt"
	"sort"
	"strings"
)

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
