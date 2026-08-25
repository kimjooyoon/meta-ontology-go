package feedbackpredecessor

type useCaseFile struct {
	Schema string    `json:"schema"`
	Cases  []useCase `json:"cases"`
}

type useCase struct {
	ID                   string `json:"id"`
	Actor                string `json:"actor"`
	Trigger              string `json:"trigger"`
	Cause                string `json:"cause"`
	CandidateState       string `json:"candidate_state"`
	ExpectedDecision     string `json:"expected_decision"`
	ExpectedReason       string `json:"expected_reason"`
	ExpectedResolution   string `json:"expected_resolution"`
	ExpectedPromotion    bool   `json:"expected_promotion_authorized"`
	ExpectedReadinessBPS int    `json:"expected_readiness_bps"`
	ExpectedContinuityBPS int   `json:"expected_continuity_bps"`
	ExpectedAmbiguity    int    `json:"expected_ambiguity"`
	ExpectedWrites       int    `json:"expected_writes"`
	ExpectedSatisfied    int    `json:"expected_satisfied_indicators"`
	OperatorAction       string `json:"operator_action"`
}
