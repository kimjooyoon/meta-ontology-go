package cycles

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// ContractVersion identifies the detector input and evidence contract.
const ContractVersion = "cycles/detection/v1"

// HostStage identifies the implementation that produced an observation.
type HostStage string

const (
	GoHostedStage   HostStage = "go-hosted"
	GoooHostedStage HostStage = "gooo-hosted"
)

// StageStatus distinguishes an observed implementation from a planned one.
type StageStatus string

const (
	StageObserved StageStatus = "observed"
	StagePending  StageStatus = "pending"
)

// Evidence is a deterministic observation of detector diagnostics. A pending
// stage deliberately has no digest and cannot be mistaken for a passing run.
type Evidence struct {
	ContractVersion string
	Stage           HostStage
	Status          StageStatus
	Producer        string
	Codes           []Code
	Digest          string
}

// NewEvidence records an observed run without treating its diagnostic result
// as success. The digest covers the ordered diagnostic codes and messages.
func NewEvidence(stage HostStage, producer string, diagnostics Diagnostics) Evidence {
	canonical := diagnostics.Error()
	sum := sha256.Sum256([]byte(canonical))
	return Evidence{
		ContractVersion: ContractVersion, Stage: stage, Status: StageObserved,
		Producer: strings.TrimSpace(producer), Codes: diagnostics.Codes(),
		Digest: hex.EncodeToString(sum[:]),
	}
}

// PendingEvidence records a future host without inventing implementation
// evidence for it.
func PendingEvidence(stage HostStage, producer string) Evidence {
	return Evidence{ContractVersion: ContractVersion, Stage: stage,
		Status: StagePending, Producer: strings.TrimSpace(producer)}
}

// Validate checks that an evidence record does not claim an unsupported
// stage or an observed result without its deterministic digest.
func (e Evidence) Validate() error {
	if e.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported contract version %q", e.ContractVersion)
	}
	if e.Stage != GoHostedStage && e.Stage != GoooHostedStage {
		return fmt.Errorf("unsupported host stage %q", e.Stage)
	}
	if strings.TrimSpace(e.Producer) == "" {
		return fmt.Errorf("evidence producer is empty")
	}
	switch e.Status {
	case StagePending:
		if e.Digest != "" || len(e.Codes) != 0 {
			return fmt.Errorf("pending evidence cannot contain an observation")
		}
	case StageObserved:
		if len(e.Digest) != sha256.Size*2 {
			return fmt.Errorf("observed evidence has no SHA-256 digest")
		}
		if _, err := hex.DecodeString(e.Digest); err != nil {
			return fmt.Errorf("observed evidence digest: %v", err)
		}
	default:
		return fmt.Errorf("unsupported evidence status %q", e.Status)
	}
	return nil
}

// HostContract compares the current Go-hosted stage with the future stage.
type HostContract struct {
	Version    string
	GoHosted   Evidence
	GoooHosted Evidence
}

// CurrentHostContract returns the truthful baseline: Go-hosted is observable
// now, while gooo-hosted remains explicitly pending.
func CurrentHostContract() HostContract {
	return HostContract{
		Version:    ContractVersion,
		GoHosted:   NewEvidence(GoHostedStage, "go-hosted", nil),
		GoooHosted: PendingEvidence(GoooHostedStage, "gooo-hosted"),
	}
}

// Validate checks both stages and prevents a future stage from being silently
// represented as a successful observation.
func (c HostContract) Validate() error {
	if c.Version != ContractVersion {
		return fmt.Errorf("unsupported host contract version %q", c.Version)
	}
	if c.GoooHosted.Status != StagePending {
		return fmt.Errorf("gooo-hosted stage must remain pending until implemented")
	}
	if err := c.GoooHosted.Validate(); err != nil {
		return err
	}
	if err := c.GoHosted.Validate(); err != nil {
		return err
	}
	return validateObservedMetadata(c.GoHosted)
}

func validateObservedMetadata(evidence Evidence) error {
	if evidence.ContractVersion != ContractVersion || evidence.Stage != GoHostedStage {
		return fmt.Errorf("go-hosted evidence does not match the contract")
	}
	if strings.TrimSpace(evidence.Producer) == "" {
		return fmt.Errorf("go-hosted evidence producer is empty")
	}
	return nil
}

// EquivalentEvidence compares host-independent observations. A pending
// record is never equivalent to an observed record.
func EquivalentEvidence(left, right Evidence) bool {
	if left.Validate() != nil || right.Validate() != nil {
		return false
	}
	if left.Status != StageObserved || right.Status != StageObserved {
		return false
	}
	if left.ContractVersion != right.ContractVersion || left.Digest != right.Digest {
		return false
	}
	if len(left.Codes) != len(right.Codes) {
		return false
	}
	for i := range left.Codes {
		if left.Codes[i] != right.Codes[i] {
			return false
		}
	}
	return true
}
