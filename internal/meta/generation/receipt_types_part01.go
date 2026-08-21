package generation

const OperationReceiptSchemaVersion = "gooo/meta-operation-receipt/v1"
const ReceiptReportSchemaVersion = "gooo/meta-operation-receipt-report/v1"

type ReceiptDecision string

const ReceiptDecisionConformant ReceiptDecision = "CONFORMANT"
const ReceiptDecisionFixedPoint ReceiptDecision = "FIXED_POINT"
const ReceiptDecisionUnknown ReceiptDecision = "UNKNOWN"
const ReceiptDecisionRejected ReceiptDecision = "REJECTED"

type ReceiptReason string

const ReceiptReasonVerified ReceiptReason = "RECEIPTS_VERIFIED"
const ReceiptReasonExactFixedPoint ReceiptReason = "EXACT_FIXED_POINT"
const ReceiptReasonInvalidPlan ReceiptReason = "INVALID_GENERATION_PLAN"
const ReceiptReasonPlanNotExecutable ReceiptReason = "PLAN_NOT_EXECUTABLE"
const ReceiptReasonSetMismatch ReceiptReason = "RECEIPT_SET_MISMATCH"
const ReceiptReasonMissingIndicator ReceiptReason = "MISSING_REQUIRED_INDICATOR"
const ReceiptReasonUnknownIndicator ReceiptReason = "UNKNOWN_INDICATOR_EVIDENCE"
const ReceiptReasonRejectedIndicator ReceiptReason = "INDICATOR_REJECTED"

type IndicatorVerdict string

const IndicatorVerdictPass IndicatorVerdict = "PASS"
const IndicatorVerdictFail IndicatorVerdict = "FAIL"
const IndicatorVerdictUnknown IndicatorVerdict = "UNKNOWN"
