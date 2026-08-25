package feedbackpredecessor

const Schema = "gooo/meta-feedback-predecessor-selection/v2"

const (
	DecisionSelected   = "SELECTED"
	DecisionLower      = "LOWER_RESOLUTION"
	DecisionFailClosed = "FAIL_CLOSED"
)

const (
	ReasonSelected         = "PREDECESSOR_FEEDBACK_SELECTED"
	ReasonNotFound         = "PREDECESSOR_FEEDBACK_NOT_FOUND"
	ReasonCanonicalUnbound = "PREDECESSOR_CANONICAL_RUN_UNBOUND"
	ReasonUnsuccessful     = "PREDECESSOR_CANONICAL_RUN_UNSUCCESSFUL"
	ReasonUnavailable      = "PREDECESSOR_FEEDBACK_ARTIFACT_UNAVAILABLE"
	ReasonReceiptUnbound   = "PREDECESSOR_FEEDBACK_RECEIPT_UNBOUND"
	ReasonAmbiguous        = "PREDECESSOR_FEEDBACK_AMBIGUOUS"
	ReasonWriteEffect      = "PREDECESSOR_FEEDBACK_WRITE_EFFECT"
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
