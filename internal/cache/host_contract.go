package cache

import (
	"errors"
	"fmt"
	"strings"
)

// HostStage identifies the implementation that produced a cache projection.
// Stage names are part of the cache contract so equivalent inputs from two
// hosts cannot silently share an entry.
type HostStage string

const (
	// GoHostedStage is the current bootstrap implementation.
	GoHostedStage HostStage = "go-hosted"
	// GoooHostedStage is the planned self-hosted implementation.
	GoooHostedStage HostStage = "gooo-hosted"
	// DefaultHostStage keeps existing callers on the current implementation.
	DefaultHostStage = GoHostedStage
)

// EvidenceStatus describes what a host-stage record proves. Deferred and
// blocked are intentionally distinct from verified and are never successes.
type EvidenceStatus string

const (
	EvidenceVerified EvidenceStatus = "verified"
	EvidenceDeferred EvidenceStatus = "deferred"
	EvidenceBlocked  EvidenceStatus = "blocked"
)

var (
	ErrInvalidHostStage     = errors.New("invalid cache host stage")
	ErrInvalidStageEvidence = errors.New("invalid cache stage evidence")
	ErrUnimplementedStage   = errors.New("cache host stage is not implemented")
)

// StageEvidence is a small, comparable contract record for self-hosting
// evidence. A future host may be named now, but it cannot claim verification
// until its implementation and independent checks exist.
type StageEvidence struct {
	Stage     HostStage
	Status    EvidenceStatus
	Authority string
}

// Validate checks the stage contract without treating a future stage as live.
func (e StageEvidence) Validate() error {
	if !e.Stage.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidStageEvidence, e.Stage)
	}
	if e.Status != EvidenceVerified && e.Status != EvidenceDeferred && e.Status != EvidenceBlocked {
		return fmt.Errorf("%w: unknown status %q", ErrInvalidStageEvidence, e.Status)
	}
	if strings.TrimSpace(e.Authority) == "" {
		return fmt.Errorf("%w: authority is empty", ErrInvalidStageEvidence)
	}
	if e.Stage == GoooHostedStage && e.Status == EvidenceVerified {
		return fmt.Errorf("%w: %s", ErrUnimplementedStage, e.Stage)
	}
	return nil
}

// CurrentStageEvidence reports the verified bootstrap boundary.
func CurrentStageEvidence() StageEvidence {
	return StageEvidence{Stage: GoHostedStage, Status: EvidenceVerified,
		Authority: "Go verifier is authoritative at self-hosting Stage 0"}
}

// FutureStageEvidence records the planned host without claiming it works.
func FutureStageEvidence() StageEvidence {
	return StageEvidence{Stage: GoooHostedStage, Status: EvidenceDeferred,
		Authority: "gooo-hosted implementation and parity checks are not available"}
}

// Valid reports whether s is one of the two declared host stages.
func (s HostStage) Valid() bool {
	return s == GoHostedStage || s == GoooHostedStage
}

// String returns the stable serialized stage name.
func (s HostStage) String() string { return string(s) }
