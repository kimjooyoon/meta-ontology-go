package predecessorselection

const (
	Schema             = "gooo/language-readiness-predecessor-selection/v2"
	DecisionSelected   = "SELECTED"
	DecisionFailClosed = "FAIL_CLOSED"
	ProducerJobName    = "language-concept-artifact"
)

const (
	ReasonSelected             = "READINESS_PREDECESSOR_SELECTED"
	ReasonNotFound             = "READINESS_PREDECESSOR_NOT_FOUND"
	ReasonUnbound              = "READINESS_PREDECESSOR_CANONICAL_RUN_UNBOUND"
	ReasonFailed               = "READINESS_PREDECESSOR_RUN_FAILED"
	ReasonProducer             = "READINESS_PREDECESSOR_PRODUCER_NOT_CONFORMANT"
	ReasonExpired              = "READINESS_PREDECESSOR_ARTIFACT_EXPIRED"
	ReasonInvalid              = "READINESS_PREDECESSOR_PAYLOAD_INVALID"
	ReasonAmbiguous            = "READINESS_PREDECESSOR_AMBIGUOUS"
	ReasonWriteEffect          = "READINESS_PREDECESSOR_WRITE_EFFECT"
	ReasonHTTPFailure          = "WORKFLOW_RUN_HTTP_FAILURE"
	ReasonPaginationIncomplete = "WORKFLOW_RUN_PAGINATION_INCOMPLETE"
	ReasonPaginationRepeated   = "WORKFLOW_RUN_NEXT_LINK_REPEATED"
	ReasonPaginationCap        = "WORKFLOW_RUN_PAGE_CAP_EXCEEDED"
	ReasonPaginationMalformed  = "WORKFLOW_RUN_LINK_MALFORMED"
	ReasonPaginationDuplicate  = "WORKFLOW_RUN_DUPLICATE_ID"
	ReasonResponseMalformed    = "WORKFLOW_RUN_RESPONSE_MALFORMED"
)
