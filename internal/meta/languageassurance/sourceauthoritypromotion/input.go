package sourceauthoritypromotion

type assuranceDocument struct {
	Schema, SubjectSHA, DenominatorID, DenominatorDigest string
	AssuranceDecision, CandidateDecision                  string
	Denominator                                           []assuranceDefinition
	Obligations                                           []assuranceObligation
	Summary                                               assuranceSummary
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
	RepositoryWrites                                                       int
}

type upstreamDocument struct {
	Schema, SubjectSHA, Decision, Resolution                 string
	DenominatorID, DenominatorDigest                         string
	RepositoryWrites, PromotionCreditBPS                     int
	Summary                                                  upstreamSummary
	Cases                                                    []upstreamCase
}

type upstreamSummary struct {
	CasesTotal, CasesPassed, ExactAllow, FailClosed, CoverageBPS int
}

type upstreamCase struct {
	ID, ExpectedObservation, ExpectedResolution string
	ExpectedEnforcement, ExpectedReason          string
	Passed                                       bool
	Receipt                                      upstreamReceipt
}

type upstreamReceipt struct {
	SubjectSHA, Observation, Resolution, Enforcement, Reason string
	RepositoryWrites, PromotionCreditBPS                     int
	Snapshot                                                  *upstreamSnapshot
	Indicators                                                []upstreamIndicator
}

type upstreamSnapshot struct {
	Digest, SourceRef, AuthorityRef string
	Bytes                            int
	Authority                        upstreamAuthority
	Selection                        upstreamSelection
}

type upstreamAuthority struct{ Repository, Revision, Path string }
type upstreamSelection struct{ StartLine, EndLine int }
type upstreamIndicator struct {
	Class, ProofChoice string
	Satisfied          bool
}
