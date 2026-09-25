package proposalcompat

const (
	Schema             = "gooo/proposal-promotion-compatibility/v1"
	LegacySchema       = "gooo/autonomous-change-proposal-promotion/v1"
	DecisionPass       = "PASS"
	DecisionFailClosed = "FAIL_CLOSED"
	ReasonReady        = "PROMOTION_COMPATIBILITY_PROJECTION_READY"
	ReasonRejected     = "PROMOTION_COMPATIBILITY_PROJECTION_REJECTED"
	totalCoordinates   = 6
	projectedFields    = 8
)
