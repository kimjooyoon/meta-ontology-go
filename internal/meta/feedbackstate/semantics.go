package feedbackstate

const (
	decisionFixed   = "FIXED_POINT"
	decisionImprove = "IMPROVE"
	decisionLower   = "LOWER_RESOLUTION"
	decisionClosed  = "FAIL_CLOSED"
)

type semanticResult struct {
	snapshot   Snapshot
	valid      bool
	reason     string
	falseFixed int
}

func resolve(report archivedReport, receiptDigest string) semanticResult {
	snapshot := Snapshot{
		SourceDecision: report.SourceDecision, Decision: report.Decision, Reason: report.Reason,
		FromResolution: report.FromResolution, ToResolution: report.ToResolution,
		NextOperation: report.Feedback.NextOperation, PreviousDescents: report.PreviousDescents,
		Descents: report.Descents, ReceiptDigest: receiptDigest, ReportDigest: report.ReportDigest,
	}
	level, knownResolution := resolutionIndex(report.FromResolution)
	same := knownResolution && report.FromResolution == report.ToResolution &&
		report.PreviousDescents == level && report.Descents == level
	if report.Reason == "" {
		return semanticResult{snapshot, false, "FEEDBACK_SEMANTIC_REASON_MISSING", 0}
	}
	switch report.Decision {
	case decisionFixed:
		valid := report.SourceDecision == decisionFixed && report.Feedback.Decision == decisionFixed && same
		if !valid {
			return semanticResult{snapshot, false, "FALSE_FIXED_POINT_REJECTED", 1}
		}
		snapshot.NextOperation = "none"
		return semanticResult{snapshot: snapshot, valid: true}
	case decisionImprove:
		valid := report.SourceDecision == decisionImprove && report.Feedback.Decision == decisionImprove && same && snapshot.NextOperation != ""
		if !valid {
			return semanticResult{snapshot, false, "IMPROVEMENT_OPERATION_UNBOUND", 0}
		}
		return semanticResult{snapshot: snapshot, valid: true}
	case decisionLower, decisionClosed:
		return resolveFailure(snapshot, report)
	default:
		return semanticResult{snapshot, false, "FEEDBACK_SEMANTIC_DECISION_UNKNOWN", 0}
	}
}

func resolveFailure(snapshot Snapshot, report archivedReport) semanticResult {
	from, fromOK := resolutionIndex(report.FromResolution)
	to, toOK := resolutionIndex(report.ToResolution)
	boundFailure := report.SourceDecision == decisionClosed && report.Feedback.Decision == decisionClosed
	if !boundFailure {
		return semanticResult{snapshot, false, "FAILURE_DECISION_UNBOUND", 0}
	}
	if fromOK && toOK && to == from+1 && report.PreviousDescents == from && report.Descents == to {
		snapshot.Decision = decisionLower
		snapshot.NextOperation = "re-evaluate-at-" + report.ToResolution
		return semanticResult{snapshot: snapshot, valid: true}
	}
	if fromOK && from == 2 && report.ToResolution == report.FromResolution && report.PreviousDescents == 2 && report.Descents == 2 {
		snapshot.Decision, snapshot.NextOperation = decisionClosed, "halt"
		return semanticResult{snapshot: snapshot, valid: true}
	}
	return semanticResult{snapshot, false, "SEMANTIC_RESOLUTION_TRANSITION_INVALID", 0}
}

func resolutionIndex(value string) (int, bool) {
	for index, candidate := range []string{"exact_operation", "operation_class", "invariant_only"} {
		if value == candidate {
			return index, true
		}
	}
	return 0, false
}
