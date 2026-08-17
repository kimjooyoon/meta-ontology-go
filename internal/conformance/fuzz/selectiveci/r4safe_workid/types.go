// Package r4safe_workid is a research-only audit of the legacy WorkID rule.
package r4safe_workid

type LegacyWorkID string
type AuditDigest string

type Decision uint8

const (
	DecisionUnknown    Decision = 0
	DecisionPass       Decision = 1
	DecisionFailClosed Decision = 2
)

type Reason uint8

const (
	ReasonRequiredInputMissing   Reason = 0
	ReasonDerived                Reason = 1
	ReasonMatchingCallerOverride Reason = 2
	ReasonMalformedOverride      Reason = 3
	ReasonWorkIDMismatch         Reason = 4
)

type EnforcementEffect uint8

const EnforcementNoEffect EnforcementEffect = 0

type Input struct {
	SnapshotDigest string
	ObligationID   string
	PathID         string
	PolicyDigest   string
	CallerWorkID   string
}

type Result struct {
	Decision            Decision
	Reason              Reason
	LegacyWorkID        LegacyWorkID
	FullSuiteRequired   bool
	ExecutionAuthorized bool
	EnforcementEffect   EnforcementEffect
	CanonicalDigest     AuditDigest
}
