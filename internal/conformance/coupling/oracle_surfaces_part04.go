package coupling

func validateReceiptClaim(receipt CouplingReceipt, beforeDigest, afterDigest, deltaText string) oracleValidation {
	switch receipt.ChangeClaim {
	case ClaimDelta:
		if receipt.ReceiptKind != ReceiptSemanticDelta || beforeDigest == afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.SemanticDelta != deltaText || receipt.SemanticDelta == "" || receipt.SemanticDeltaDigest != digestBytes([]byte(deltaText)) {
			return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
		}
		if receipt.AuthoritativeSourceRef == "" || receipt.AuthoritySourceAfterDigest == "" {
			return oracleValidation{DecisionFailClosed, ReasonDeltaWithoutSource}
		}
	case ClaimNoDelta:
		if receipt.ReceiptKind != ReceiptNoSemanticDelta || beforeDigest != afterDigest {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
		if receipt.SemanticDelta != "" || receipt.SemanticDeltaDigest != "" || receipt.AuthoritativeSourceRef != "" {
			return oracleValidation{DecisionFailClosed, ReasonNoDeltaWithoutEquality}
		}
	default:
		return oracleValidation{DecisionFailClosed, ReasonInvalidDelta}
	}
	if len(receipt.EvidenceRefs) == 0 || len(receipt.OriginPathIDs) == 0 || receipt.ClaimRecordID == "" {
		return oracleValidation{DecisionFailClosed, ReasonPathMalformed}
	}
	return oracleValidation{}
}
