package valueexecution

// SelfImprovementEvidenceLedgerStatus describes a durable evidence decision,
// not a permission to promote or deploy a candidate.
type SelfImprovementLedgerDecision string

const (
	SelfImprovementLedgerDecisionUnknown              SelfImprovementLedgerDecision = "UNKNOWN"
	SelfImprovementLedgerDecisionInsufficientEvidence SelfImprovementLedgerDecision = "INSUFFICIENT_EVIDENCE"
	SelfImprovementLedgerDecisionBlockedRegression    SelfImprovementLedgerDecision = "BLOCKED_REGRESSION"
	SelfImprovementLedgerDecisionPromotionEligible    SelfImprovementLedgerDecision = "PROMOTION_ELIGIBLE"
)

const SelfImprovementEvidenceLedgerSchema = "gooo/self-improvement-evidence-ledger/v1"

// SelfImprovementEvidenceLedger is append-only from the caller's perspective.
// Earlier UNKNOWN and regression observations remain in the digest and cannot
// be erased by a later improvement observation.
type SelfImprovementEvidenceLedger struct {
	Schema               string                               `json:"schema"`
	RequiredObservations int                                  `json:"required_observations"`
	Observations         []ExecutedSelfImprovementObservation `json:"observations"`
	NonAuthorizing       bool                                 `json:"non_authorizing"`
	Digest               string                               `json:"digest"`
}

// SelfImprovementLedgerObservation summarizes the complete ledger while
// retaining counts for audit and future counterexample analysis.
type SelfImprovementLedgerObservation struct {
	Schema               string                        `json:"schema"`
	RequiredObservations int                           `json:"required_observations"`
	Observed             int                           `json:"observed"`
	Improved             int                           `json:"improved"`
	Regressed            int                           `json:"regressed"`
	Unchanged            int                           `json:"unchanged"`
	Unknown              int                           `json:"unknown"`
	Decision             SelfImprovementLedgerDecision `json:"decision"`
	Reason               string                        `json:"reason"`
	NonAuthorizing       bool                          `json:"non_authorizing"`
	Digest               string                        `json:"digest"`
}

// NewSelfImprovementEvidenceLedger starts a ledger with a minimum repeated
// evidence requirement. A non-positive requirement is fail-closed to one.
func NewSelfImprovementEvidenceLedger(requiredObservations int) SelfImprovementEvidenceLedger {
	if requiredObservations < 1 {
		requiredObservations = 1
	}
	ledger := SelfImprovementEvidenceLedger{
		Schema:               SelfImprovementEvidenceLedgerSchema,
		RequiredObservations: requiredObservations,
		Observations:         []ExecutedSelfImprovementObservation{},
		NonAuthorizing:       true,
	}
	ledger.Digest = selfImprovementEvidenceLedgerDigest(ledger)
	return ledger
}

// Append returns a new ledger and does not mutate the prior observation list.
func (ledger SelfImprovementEvidenceLedger) Append(observation ExecutedSelfImprovementObservation) SelfImprovementEvidenceLedger {
	if ledger.Schema == "" {
		ledger.Schema = SelfImprovementEvidenceLedgerSchema
	}
	if ledger.RequiredObservations < 1 {
		ledger.RequiredObservations = 1
	}
	observations := append([]ExecutedSelfImprovementObservation(nil), ledger.Observations...)
	ledger.Observations = append(observations, observation)
	ledger.NonAuthorizing = true
	ledger.Digest = selfImprovementEvidenceLedgerDigest(ledger)
	return ledger
}

// Observe evaluates every recorded entry. UNKNOWN evidence takes precedence
// over eligibility, and any regression blocks eligibility without deleting
// the improvement count or the original counterexample.
func (ledger SelfImprovementEvidenceLedger) Observe() SelfImprovementLedgerObservation {
	observation := SelfImprovementLedgerObservation{
		Schema:               SelfImprovementEvidenceLedgerSchema,
		RequiredObservations: ledger.RequiredObservations,
		Observed:             len(ledger.Observations),
		Decision:             SelfImprovementLedgerDecisionUnknown,
		NonAuthorizing:       true,
	}
	if ledger.Schema != SelfImprovementEvidenceLedgerSchema || ledger.RequiredObservations < 1 || !ledger.NonAuthorizing || !validDigest(ledger.Digest) {
		observation.Reason = "LEDGER_IDENTITY_UNKNOWN"
		observation.Digest = selfImprovementLedgerObservationDigest(observation)
		return observation
	}
	for _, entry := range ledger.Observations {
		switch {
		case !entry.NonAuthorizing || !validDigest(entry.Digest) || entry.Status == SelfImprovementStatusUnknown:
			observation.Unknown++
		case entry.Status == SelfImprovementStatusImproved:
			observation.Improved++
		case entry.Status == SelfImprovementStatusRegressed:
			observation.Regressed++
		case entry.Status == SelfImprovementStatusUnchanged:
			observation.Unchanged++
		default:
			observation.Unknown++
		}
	}
	switch {
	case observation.Unknown > 0:
		observation.Reason = "LEDGER_CONTAINS_UNKNOWN_EVIDENCE"
	case observation.Regressed > 0:
		observation.Decision = SelfImprovementLedgerDecisionBlockedRegression
		observation.Reason = "LEDGER_CONTAINS_REGRESSION"
	case observation.Improved < observation.RequiredObservations:
		observation.Decision = SelfImprovementLedgerDecisionInsufficientEvidence
		observation.Reason = "LEDGER_REQUIRES_REPEATED_IMPROVEMENT"
	default:
		observation.Decision = SelfImprovementLedgerDecisionPromotionEligible
		observation.Reason = "LEDGER_REPEATED_IMPROVEMENT_BOUND"
	}
	observation.Digest = selfImprovementLedgerObservationDigest(observation)
	return observation
}

func selfImprovementEvidenceLedgerDigest(ledger SelfImprovementEvidenceLedger) string {
	ledger.Digest = ""
	return digestValue(ledger)
}

func selfImprovementLedgerObservationDigest(observation SelfImprovementLedgerObservation) string {
	observation.Digest = ""
	return digestValue(observation)
}
