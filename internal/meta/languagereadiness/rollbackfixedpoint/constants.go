package rollbackfixedpoint

const (
	Schema             = "gooo/language-rollback-fixed-point/v1"
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ModeRecovered      = "RECOVERED_FIXED_POINT"
	ModeAuthorized     = "PROMOTION_AUTHORIZED"
	ResolutionExact    = "EXACT"
	ResolutionLower    = "LOWER_RESOLUTION"
	ReasonRecovered    = "ROLLBACK_FIXED_POINT_RECOVERED"
	ReasonAuthorized   = "PROMOTION_ALREADY_AUTHORIZED"
	ReasonUnknown      = "ROLLBACK_EVIDENCE_UNKNOWN"
	ReasonRejected     = "ROLLBACK_FIXED_POINT_REJECTED"
	totalCoordinates   = 10
)
