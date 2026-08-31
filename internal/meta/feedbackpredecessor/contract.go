package feedbackpredecessor

const Schema = "gooo/meta-feedback-predecessor-selection/v2"

const (
	DecisionSelected   = "SELECTED"
	DecisionFoundation = "FOUNDATION"
	DecisionRefuted    = "REFUTED"
	DecisionLower      = "LOWER_RESOLUTION"
	DecisionFailClosed = "FAIL_CLOSED"
)

const (
	ReasonSelected             = "PREDECESSOR_FEEDBACK_SELECTED"
	ReasonNotFound             = "PREDECESSOR_FEEDBACK_NOT_FOUND"
	ReasonCanonicalUnbound     = "PREDECESSOR_CANONICAL_RUN_UNBOUND"
	ReasonUnsuccessful         = "PREDECESSOR_CANONICAL_RUN_UNSUCCESSFUL"
	ReasonUnavailable          = "PREDECESSOR_FEEDBACK_ARTIFACT_UNAVAILABLE"
	ReasonReceiptUnbound       = "PREDECESSOR_FEEDBACK_RECEIPT_UNBOUND"
	ReasonAmbiguous            = "PREDECESSOR_FEEDBACK_AMBIGUOUS"
	ReasonWriteEffect          = "PREDECESSOR_FEEDBACK_WRITE_EFFECT"
	ReasonFoundationRegression = "PREDECESSOR_CHAIN_BROKEN_BY_CONFIRMED_REGRESSION"
	ReasonFoundationRefuted    = "FOUNDATION_RECOVERY_REFUTED"
)

const (
	FoundationProofChoice                               = "FOUNDATION"
	FoundationNextOperation                             = "RESTORE_NORMAL_PREDECESSOR_CHAIN"
	FoundationLastKnownGoodSHA                          = "bc5dc21788aa4c7d46d1f8ab516f8218bb423fdc"
	FoundationMissingPredecessorSHA                     = "cd9727af80f5118405290d3be96890c18e1529c0"
	FoundationLastKnownGoodRunID                  int64 = 32572203736
	FoundationLastKnownGoodArtifactID             int64 = 9475640134
	FoundationLastKnownGoodArtifactName                 = "artifact-feedback-resolution-bc5dc21788aa4c7d46d1f8ab516f8218bb423fdc"
	FoundationLastKnownGoodArtifactDigest               = "sha256:7741a5fdb5c304715f8b8a330264e1a8bb9f3c10760dbffbe3e7d9a8f247c944"
	FoundationLastKnownGoodReceiptDigest                = "sha256:a4e36893e070c1dd01284c37c643daaf1106536be08e94c0fd4e236789bee19c"
	FoundationLastKnownGoodFeedbackReportDigest         = "sha256:c58242d820ce514e8f9ad839b0dcf222194f4640217f537cafc9ee0661e0135e"
	FoundationLastKnownGoodResolutionReportDigest       = "sha256:b8f2d57f1a0c02ee172c617b1557a243ecb5470786779b68efac1cf6a836fd86"
	FoundationLastKnownGoodFeedbackReason               = "NEXT_CYCLE_FEEDBACK_FIXED_POINT"
)

const (
	ResolutionExact     = "exact_operation"
	ResolutionClass     = "operation_class"
	ResolutionInvariant = "invariant_only"

	OperationConsume    = "consume-predecessor-semantic-state"
	OperationReevaluate = "re-evaluate-predecessor-at-exact-operation"
	OperationHalt       = "halt"
)

const (
	ClassOutcome   = "outcome"
	ClassDriver    = "driver"
	ClassGuardrail = "guardrail"

	RelationGreaterOrEqual = "greater_or_equal"
	RelationLessOrEqual    = "less_or_equal"

	ProofFoundation = "foundation"
	ProofCoherence  = "coherence"
	ProofRegression = "regression"
)
