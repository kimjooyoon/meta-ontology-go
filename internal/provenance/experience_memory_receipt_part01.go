package provenance

import (
	"fmt"
	"strings"
)

const ExperienceMemoryReceiptSchema = "gooo/experience-memory-receipt/v1"

type ExperienceMemoryObservationStatus string

const (
	ExperienceMemoryClosed  ExperienceMemoryObservationStatus = "CLOSED"
	ExperienceMemoryUnknown ExperienceMemoryObservationStatus = "UNKNOWN"
	ExperienceMemoryRefuted ExperienceMemoryObservationStatus = "REFUTED"
)

type ExperienceMemoryFingerprint struct {
	ScenarioID      string
	SourceDigest    string
	FixtureDigest   string
	ToolchainDigest string
	EvaluatorDigest string
	ScopeDigest     string
}

type ExperienceMemoryReceipt struct {
	SchemaVersion string
	ExperienceID  string
	Fingerprint   ExperienceMemoryFingerprint
	OutcomeDigest string
	Status        ExperienceMemoryObservationStatus

	CounterexamplePreserved bool
	MissingStageIndex       int
	MissingStage            string
	UnknownReason           string
}

func NewExperienceMemoryReceipt() ExperienceMemoryReceipt {
	return ExperienceMemoryReceipt{
		SchemaVersion:     ExperienceMemoryReceiptSchema,
		MissingStageIndex: -1,
	}
}

func (r *ExperienceMemoryReceipt) PreserveFirstUnknown(stageIndex int, stage, reason string) error {
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

func (r ExperienceMemoryReceipt) Observe(expected ExperienceMemoryFingerprint) ExperienceMemoryReceipt {
	if r.Status == "" {
		r.Status = ExperienceMemoryUnknown
	}
	if r.Fingerprint != expected {
		_ = r.PreserveFirstUnknown(0, "fingerprint", "experience memory fingerprint is stale or incomparable")
		r.Status = ExperienceMemoryUnknown
		return r
	}
	if strings.TrimSpace(r.OutcomeDigest) == "" {
		_ = r.PreserveFirstUnknown(1, "outcome", "experience memory has no immutable outcome receipt")
		r.Status = ExperienceMemoryUnknown
	}
	return r
}

func (r ExperienceMemoryReceipt) Validate() error {
	if r.SchemaVersion != ExperienceMemoryReceiptSchema {
		return fmt.Errorf("unexpected experience memory schema %q", r.SchemaVersion)
	}
	if strings.TrimSpace(r.ExperienceID) == "" {
		return fmt.Errorf("experience id must be non-empty")
	}
	if strings.TrimSpace(r.Fingerprint.ScenarioID) == "" {
		return fmt.Errorf("scenario id must be non-empty")
	}
	if strings.TrimSpace(r.Fingerprint.SourceDigest) == "" ||
		strings.TrimSpace(r.Fingerprint.FixtureDigest) == "" ||
		strings.TrimSpace(r.Fingerprint.ToolchainDigest) == "" ||
		strings.TrimSpace(r.Fingerprint.EvaluatorDigest) == "" ||
		strings.TrimSpace(r.Fingerprint.ScopeDigest) == "" {
		return fmt.Errorf("experience fingerprint must be complete")
	}
	switch r.Status {
	case ExperienceMemoryClosed, ExperienceMemoryUnknown, ExperienceMemoryRefuted:
	default:
		return fmt.Errorf("invalid experience memory status %q", r.Status)
	}
	if r.Status == ExperienceMemoryClosed && strings.TrimSpace(r.OutcomeDigest) == "" {
		return fmt.Errorf("CLOSED experience memory requires immutable outcome receipt")
	}
	if r.Status == ExperienceMemoryUnknown {
		if r.MissingStageIndex < 0 || strings.TrimSpace(r.MissingStage) == "" || strings.TrimSpace(r.UnknownReason) == "" {
			return fmt.Errorf("UNKNOWN experience memory requires first missing stage coordinates")
		}
	}
	return nil
}
