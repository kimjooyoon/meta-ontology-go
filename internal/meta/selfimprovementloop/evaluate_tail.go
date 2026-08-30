package selfimprovementloop

import "strings"

func evaluateReceipt(in Input) cellEvaluation {
	if in.Receipt.Captured && validDigest(in.Receipt.Digest) {
		return cellEvaluation{Decision: DecisionClosed, Reason: "RECEIPT_CAPTURED"}
	}
	return cellEvaluation{Decision: DecisionUnknown, Reason: "RECEIPT_NOT_CAPTURED", Unknown: unknown(
		"CAPTURE_RECEIPT", "capture", "receipt digest is absent", "MISSING_RECEIPT", "CAPTURE_RECEIPT", "receipt",
	)}
}

func pairCell(pair pairEvaluation) cellEvaluation {
	if pair.Decision == DecisionUnknown {
		return cellEvaluation{Decision: pair.Decision, Reason: pair.Reason, Unknown: unknown(
			"COMPARE_EXACT_PAIR", "match-before-after", pair.Reason, "MISSING_EXACT_INTEGER_PAIR", "CAPTURE_RECEIPT", "before_after_pair",
		)}
	}
	return cellEvaluation{Decision: pair.Decision, Reason: pair.Reason}
}

func evaluateHuman(in Input) cellEvaluation {
	switch strings.ToUpper(strings.TrimSpace(in.Human.Decision)) {
	case "APPROVE":
		return cellEvaluation{Decision: DecisionClosed, Reason: "HUMAN_APPROVED"}
	case "REJECT":
		return cellEvaluation{Decision: DecisionRefuted, Reason: "HUMAN_REJECTED"}
	default:
		return cellEvaluation{Decision: DecisionUnknown, Reason: "HUMAN_DECISION_ABSENT", Unknown: unknown(
			"HUMAN_DECISION", "decide", "human decision is absent", "MISSING_HUMAN_DECISION", "HUMAN_DECISION", "human",
		)}
	}
}
