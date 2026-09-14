package feedbackstate

func Evaluate(input Input) (Report, error) {
	report := Report{Schema: ReportSchema, Repository: input.Repository, PredecessorSHA: input.PredecessorSHA}
	observed := observation{writes: input.RepositoryWrites}
	receipt, err := decode(input.Receipt)
	if err != nil {
		report.Decision, report.Reason = decisionClosed, "PREDECESSOR_RECEIPT_MALFORMED"
		return finish(report, observed), nil
	}
	observed.writes, err = AggregateRepositoryWrites(input.RepositoryWrites, receipt.RepositoryWrites, receipt.Report.RepositoryWrites)
	if err != nil {
		return Report{}, err
	}
	observed.identity = receipt.Schema == ReceiptSchema && receipt.Report.Schema == ResolutionSchema &&
		receipt.Report.Feedback.Repository == input.Repository && receipt.Report.Feedback.CommitSHA == input.PredecessorSHA &&
		input.Selection.ArtifactID > 0 && input.Selection.RunID > 0 && input.Selection.RunAttempt > 0
	observed.payload = input.PayloadDigest == payloadDigest(input.Receipt)
	observed.replay = receipt.ReplayVerified && receipt.ReplayReportDigest != "" &&
		receipt.ReplayReportDigest == receipt.Report.ReportDigest
	semantic := resolve(receipt.Report, receipt.ReceiptDigest)
	observed.semantic, observed.falseFixed = semantic.valid, semantic.falseFixed
	observed.descents = receipt.Report.Descents
	reason := failureReason(input, receipt, observed, semantic.reason)
	if reason != "" {
		report.Decision, report.Reason = decisionClosed, reason
		return finish(report, observed), nil
	}
	report.Decision, report.Reason = "READY", "PREDECESSOR_SEMANTIC_SNAPSHOT_READY"
	report.Snapshot = &semantic.snapshot
	return finish(report, observed), nil
}

func failureReason(input Input, receipt archivedReceipt, observed observation, semanticReason string) string {
	if !observed.identity {
		return "PREDECESSOR_RECEIPT_IDENTITY_MISMATCH"
	}
	if input.Selection.ReceiptDigest == "" || input.Selection.ReceiptDigest != receipt.ReceiptDigest {
		return "PREDECESSOR_RECEIPT_BINDING_MISMATCH"
	}
	if !observed.payload {
		return "PREDECESSOR_PAYLOAD_DIGEST_MISMATCH"
	}
	if !observed.replay {
		return "PREDECESSOR_REPLAY_UNVERIFIED"
	}
	if observed.writes != 0 {
		return "PREDECESSOR_WRITE_EFFECT"
	}
	if !observed.semantic {
		return semanticReason
	}
	return ""
}
