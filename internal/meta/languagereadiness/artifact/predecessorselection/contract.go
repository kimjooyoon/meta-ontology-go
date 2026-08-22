package predecessorselection

const (
	Schema             = "gooo/language-readiness-predecessor-selection/v1"
	DecisionSelected   = "SELECTED"
	DecisionFailClosed = "FAIL_CLOSED"
)

const (
	ReasonSelected    = "READINESS_PREDECESSOR_SELECTED"
	ReasonNotFound    = "READINESS_PREDECESSOR_NOT_FOUND"
	ReasonUnbound     = "READINESS_PREDECESSOR_CANONICAL_RUN_UNBOUND"
	ReasonFailed      = "READINESS_PREDECESSOR_RUN_FAILED"
	ReasonExpired     = "READINESS_PREDECESSOR_ARTIFACT_EXPIRED"
	ReasonInvalid     = "READINESS_PREDECESSOR_PAYLOAD_INVALID"
	ReasonAmbiguous   = "READINESS_PREDECESSOR_AMBIGUOUS"
	ReasonWriteEffect = "READINESS_PREDECESSOR_WRITE_EFFECT"
)
