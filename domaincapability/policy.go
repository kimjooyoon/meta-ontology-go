package domaincapability

// Decision is a bounded review signal, not a completeness or correctness score.
type Decision string

const (
	CollectMoreObservations Decision = "COLLECT_MORE_OBSERVATIONS"
	ReviewMissingCapabilities Decision = "REVIEW_MISSING_CAPABILITIES"
	ReviewInvestment          Decision = "REVIEW_INVESTMENT"
	NoActionableSignal        Decision = "NO_ACTIONABLE_SIGNAL"
)

// Policy makes the domain boundary and observation budget explicit.
type Policy struct {
	MinimumObservationCount int
	MaximumUnresolvedRatio  float64
	RequireEvidenceDigest   bool
}

// Measurement contains only evidence-backed domain observations.
type Measurement struct {
	ExpectedCapabilityCount  int
	ObservedCapabilityCount  int
	UnresolvedCapabilityCount int
	ObservationCount         int
	EvidenceBound            bool
}

// Decide returns a review signal without changing scope or executing implementation work.
func (p Policy) Decide(m Measurement) Decision {
	if m.ExpectedCapabilityCount <= 0 || m.ObservationCount < p.MinimumObservationCount {
		return CollectMoreObservations
	}
	if p.RequireEvidenceDigest && !m.EvidenceBound {
		return CollectMoreObservations
	}
	if m.UnresolvedCapabilityCount > 0 || m.ObservedCapabilityCount < m.ExpectedCapabilityCount {
		return ReviewMissingCapabilities
	}
	if m.ExpectedCapabilityCount > 0 && p.MaximumUnresolvedRatio >= 0 {
		ratio := float64(m.UnresolvedCapabilityCount) / float64(m.ExpectedCapabilityCount)
		if ratio > p.MaximumUnresolvedRatio {
			return ReviewMissingCapabilities
		}
	}
	return ReviewInvestment
}