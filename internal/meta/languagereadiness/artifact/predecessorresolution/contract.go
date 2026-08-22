package predecessorresolution

const (
	Schema      = "gooo/language-readiness-predecessor-resolution/v1"
	SearchLimit = 8
)

const (
	DecisionResolved   = "RESOLVED"
	DecisionFailClosed = "FAIL_CLOSED"
)

const (
	ReasonResolved  = "READINESS_ANCESTOR_RESOLVED"
	ReasonBlocked   = "READINESS_ANCESTOR_EVIDENCE_BLOCKED"
	ReasonExhausted = "READINESS_ANCESTOR_SEARCH_EXHAUSTED"
)
