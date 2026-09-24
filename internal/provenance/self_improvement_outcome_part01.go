package provenance

import (
	"fmt"
	"strings"
)

const SelfImprovementOutcomeSchema = "gooo/self-improvement-outcome/v1"

type SelfImprovementOutcomeStatus string

const (
	SelfImprovementOutcomeClosed  SelfImprovementOutcomeStatus = "CLOSED"
	SelfImprovementOutcomeUnknown SelfImprovementOutcomeStatus = "UNKNOWN"
	SelfImprovementOutcomeRefuted SelfImprovementOutcomeStatus = "REFUTED"
)

type SelfImprovementOutcomeReceipt struct {
	SchemaVersion string
	ScenarioID    string
	TrialIndex    int

	SourceDigest    string
	FixtureDigest   string
	ToolchainDigest string
	EvaluatorDigest string

	CounterexampleRecovered SelfImprovementOutcomeStatus
	ContractPreserved      SelfImprovementOutcomeStatus
	RegressionEvidence     SelfImprovementOutcomeStatus
	AdoptionAuthorized     SelfImprovementOutcomeStatus

	AdoptionAuthorityEvidenceDigest string
	AdoptionAuthorityIndependent    bool

	MissingStageIndex int
	MissingStage      string
	UnknownReason     string
}

func NewSelfImprovementOutcomeReceipt() SelfImprovementOutcomeReceipt {
	return SelfImprovementOutcomeReceipt{
		SchemaVersion:    SelfImprovementOutcomeSchema,
		MissingStageIndex: -1,
	}
}

func (r *SelfImprovementOutcomeReceipt) PreserveFirstUnknown(stageIndex int, stage, reason string) error {
	if stageIndex < 0 {
		return fmt.Errorf("missing stage index must be non-negative")
	}
	if strings.TrimSpace(stage) == "" {
		return fmt.Errorf("missing stage must be non-empty")
	}
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("unknown reason must be non-empty")
	}
	if r.MissingStageIndex >= 0 {
		return nil
	}
	r.MissingStageIndex = stageIndex
	r.MissingStage = stage
	r.UnknownReason = reason
	return nil
}

func (r SelfImprovementOutcomeReceipt) Validate() error {
	if r.SchemaVersion != SelfImprovementOutcomeSchema {
		return fmt.Errorf("unexpected outcome schema %q", r.SchemaVersion)
	}
	if strings.TrimSpace(r.ScenarioID) == "" {
		return fmt.Errorf("scenario id must be non-empty")
	}
	if r.TrialIndex < 0 {
		return fmt.Errorf("trial index must be non-negative")
	}
	for name, status := range map[string]SelfImprovementOutcomeStatus{
		"counterexample_recovered": r.CounterexampleRecovered,
		"contract_preserved":      r.ContractPreserved,
		"regression_evidence":     r.RegressionEvidence,
		"adoption_authorized":     r.AdoptionAuthorized,
	} {
		switch status {
		case SelfImprovementOutcomeClosed, SelfImprovementOutcomeUnknown, SelfImprovementOutcomeRefuted:
		default:
			return fmt.Errorf("%s has invalid status %q", name, status)
		}
	}
	if r.AdoptionAuthorized == SelfImprovementOutcomeClosed {
		if strings.TrimSpace(r.AdoptionAuthorityEvidenceDigest) == "" {
			return fmt.Errorf("adoption authority evidence is required for CLOSED adoption")
		}
		if !r.AdoptionAuthorityIndependent {
			return fmt.Errorf("adoption authority must be independent from provenance")
		}
	}
	hasUnknown := r.CounterexampleRecovered == SelfImprovementOutcomeUnknown ||
		r.ContractPreserved == SelfImprovementOutcomeUnknown ||
		r.RegressionEvidence == SelfImprovementOutcomeUnknown ||
		r.AdoptionAuthorized == SelfImprovementOutcomeUnknown
	if hasUnknown {
		if r.MissingStageIndex < 0 {
			return fmt.Errorf("UNKNOWN outcome requires missing stage index")
		}
		if strings.TrimSpace(r.MissingStage) == "" {
			return fmt.Errorf("UNKNOWN outcome requires missing stage")
		}
		if strings.TrimSpace(r.UnknownReason) == "" {
			return fmt.Errorf("UNKNOWN outcome requires reason")
		}
	}
	if r.MissingStageIndex < -1 {
		return fmt.Errorf("missing stage index must be -1 or non-negative")
	}
	return nil
}
