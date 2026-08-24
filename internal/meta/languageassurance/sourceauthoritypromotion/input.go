package sourceauthoritypromotion

type assuranceDocument struct {
	Schema, SubjectSHA, DenominatorID, DenominatorDigest string
	AssuranceDecision, CandidateDecision                 string
	Denominator                                          []assuranceDefinition `json:"denominator"`
	Obligations                                          []assuranceObligation `json:"obligations"`
	Summary                                              assuranceSummary      `json:"summary"`
}

type assuranceDefinition struct {
	MetricID, Priority, Class, ProofChoice, RequiredMetaOperation string
}

type assuranceObligation struct {
	MetricID, Status, Resolution, MetaOperation string
}

type assuranceSummary struct {
	DenominatorTotal, Operating, NotImplemented, ImplementationCoverageBPS int
	UnknownTopDecisions, UnresolvedIndicators, ViolatedGuardrails          int
	RepositoryWrites                                                       int `json:"repository_writes"`
}

type upstreamDocument struct {
	Schema, SubjectSHA, Decision, Resolution string
	DenominatorID, DenominatorDigest         string
	RepositoryWrites, PromotionCreditBPS     int
	Summary                                  upstreamSummary `json:"summary"`
	Cases                                    []upstreamCase  `json:"cases"`
}

type upstreamSummary struct {
	CasesTotal, CasesPassed, ExactAllow, FailClosed, CoverageBPS int
}

type upstreamCase struct {
	ID, ExpectedObservation, ExpectedResolution string
	ExpectedEnforcement, ExpectedReason         string
	Passed                                      bool            `json:"passed"`
	Receipt                                     upstreamReceipt `json:"receipt"`
}

type upstreamReceipt struct {
	SubjectSHA, Observation, Resolution, Enforcement, Reason string
	RepositoryWrites, PromotionCreditBPS                     int
	Snapshot                                                 *upstreamSnapshot   `json:"snapshot"`
	Indicators                                               []upstreamIndicator `json:"indicators"`
}

type upstreamSnapshot struct {
	Digest, SourceRef, AuthorityRef string
	Bytes                           int               `json:"bytes"`
	Authority                       upstreamAuthority `json:"authority"`
	Selection                       upstreamSelection `json:"selection"`
}

type upstreamAuthority struct{ Repository, Revision, Path string }
type upstreamSelection struct{ StartLine, EndLine int }
type upstreamIndicator struct {
	Class, ProofChoice string
	Satisfied          bool `json:"satisfied"`
}
