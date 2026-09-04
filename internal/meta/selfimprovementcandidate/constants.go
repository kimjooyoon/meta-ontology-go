package selfimprovementcandidate

const (
	ReportSchema       = "gooo/self-improvement-nonexecuting-candidate/v1"
	CandidateSchema    = "gooo/non-executing-improvement-candidate/v1"
	ContractID         = "gooo://self-improvement/non-executing-candidate/v1"
	PolicyVersion      = "explicit-nonclaim-priority/v1"
	DecisionProposed   = "PROPOSED"
	DecisionFailClosed = "FAIL_CLOSED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"
	indicatorTotal     = 16
	zeroDigest         = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
)

const (
	ReasonProposed        = "NON_EXECUTING_GAP_SELECTED"
	ReasonSourceUnknown   = "CANDIDATE_SOURCE_DECISION_UNKNOWN"
	ReasonSourceLowered   = "CANDIDATE_SOURCE_LOWERED"
	ReasonSourceRejected  = "CANDIDATE_SOURCE_REJECTED"
	ReasonSourceIdentity  = "CANDIDATE_SOURCE_IDENTITY_MISMATCH"
	ReasonSourceIntegrity = "CANDIDATE_SOURCE_INTEGRITY_MISMATCH"
	ReasonSourceCandidate = "CANDIDATE_SOURCE_AUTHORITY_LEAK"
	ReasonSourceAuthority = "CANDIDATE_SOURCE_EFFECT_LEAK"
	ReasonGapAbsent       = "CANDIDATE_EXPLICIT_GAP_ABSENT"
	ReasonSourceShape     = "CANDIDATE_SOURCE_SHAPE_MISMATCH"
	ReasonContractUnknown = "CANDIDATE_CONTRACT_UNKNOWN"
	ReasonContractInvalid = "CANDIDATE_CONTRACT_INVALID"
	ReasonExecutionInputUnknown = "CANDIDATE_EXECUTION_INPUT_UNKNOWN"
)
