package adapter

func validateReceiptReason(outcome ReceiptOutcome, reason RejectionKind) error {
	switch outcome {
	case ReceiptOutcomeRejected:
		if reason != "" {
			return oracleError(OracleNW003, "rejected receipt has a cancellation reason")
		}
	case ReceiptOutcomeCancelled:
		if reason != RejectionCancelled {
			return oracleError(OracleNW003, "cancelled receipt is not bound to cancellation evidence")
		}
	case ReceiptOutcomeClosed:
		if reason != RejectionClosed {
			return oracleError(OracleNW003, "closed receipt is not bound to close evidence")
		}
	}
	return nil
}
