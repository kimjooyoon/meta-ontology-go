package metainvocation

const (
	InputSchema   = "gooo/ci-plan-input/v1"
	PlanSchema    = "gooo/ci-plan/v1"
	ReceiptSchema = "gooo/ci-plan-verification-receipt/v1"
	ReportSchema  = "gooo/meta-invocation-report/v1"

	DecisionPass    = "PASS"
	DecisionClosed  = "FAIL_CLOSED"
	DecisionUnknown = "UNKNOWN"

	ResolutionExact = "EXACT"
	ResolutionLower = "LOWER_RESOLUTION"

	ClaimOpen       = "OPEN"
	ClaimDischarged = "DISCHARGED"
	ClaimRefuted    = "REFUTED"
)

type SourceCoordinate struct {
	Path        string `json:"path"`
	StartLine   int    `json:"start_line"`
	StartColumn int    `json:"start_column"`
	EndLine     int    `json:"end_line"`
	EndColumn   int    `json:"end_column"`
}
