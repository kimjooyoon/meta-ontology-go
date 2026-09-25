package assuranceeligibility

type Definition struct {
	ID                 string `json:"id"`
	ExpectedDecision   string `json:"expected_decision"`
	ExpectedResolution string `json:"expected_resolution"`
	ExpectedEffect     string `json:"expected_effect"`
	ExpectedReason     string `json:"expected_reason"`
}

type CaseResult struct {
	Definition Definition `json:"definition"`
	Passed     bool       `json:"passed"`
	Report     Report     `json:"report"`
}

type Suite struct {
	Schema            string       `json:"schema"`
	SubjectSHA        string       `json:"subject_sha"`
	DenominatorID     string       `json:"denominator_id"`
	DenominatorDigest string       `json:"denominator_digest"`
	Decision          string       `json:"decision"`
	Resolution        string       `json:"resolution"`
	Cases             []CaseResult `json:"cases"`
	Passed            int          `json:"passed"`
	Total             int          `json:"total"`
	ExactExpected     int          `json:"exact_expected"`
	UnknownExpected   int          `json:"unknown_expected"`
	InvariantExpected int          `json:"invariant_expected"`
	CoverageBPS       int          `json:"coverage_bps"`
	RepositoryWrites  int          `json:"repository_writes"`
	OfficialMutations int          `json:"official_mutations"`
	PromotionApplied  int          `json:"promotion_applied"`
	SuiteDigest       string       `json:"suite_digest"`
}

var definitions = []Definition{
	{"exact", DecisionEligible, ResolutionExact, EffectNone, ReasonEligible},
	{"missing-assurance", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-parent-report", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-parent-observation", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-parent-suite", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-capability-report", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-capability-observation", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"missing-capability-suite", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable},
	{"unknown-assurance-state", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonDecisionUnknown},
	{"unknown-parent-decision", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonDecisionUnknown},
	{"unknown-capability-decision", DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonDecisionUnknown},
	{"assurance-digest-mismatch", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonDigestMismatch},
	{"parent-false-fixed", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonParentMismatch},
	{"parent-count-mismatch", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonParentMismatch},
	{"capability-count-mismatch", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonCapabilityMismatch},
	{"capability-suite-mismatch", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonCapabilityMismatch},
	{"reference-mismatch", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonReferenceMismatch},
	{"observed-write", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonWriteObserved},
	{"official-mutation", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonMutationObserved},
	{"promotion-observed", DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonPromotionObserved},
}
