package feedbackstate

import (
	"encoding/json"
	"strings"
)

func fixture(decision, source, from, to string, previous, descents int) Input {
	receiptDigest := "sha256:" + strings.Repeat("a", 64)
	reportDigest := "sha256:" + strings.Repeat("b", 64)
	feedbackDecision, next := source, ""
	if decision == decisionImprove {
		next = "split-go-declarations"
	}
	receipt := archivedReceipt{
		Schema: ReceiptSchema, ReplayReportDigest: reportDigest, ReplayVerified: true,
		ReceiptDigest: receiptDigest,
		Report: archivedReport{
			Schema: ResolutionSchema, SourceDecision: source, Decision: decision,
			Reason: "fixture", FromResolution: from, ToResolution: to,
			PreviousDescents: previous, Descents: descents, ReportDigest: reportDigest,
			Feedback: archivedFeedback{
				CommitSHA: strings.Repeat("c", 40), Repository: "example/meta",
				Decision: feedbackDecision, NextOperation: next,
			},
		},
	}
	raw, _ := json.Marshal(receipt)
	return Input{
		Repository: "example/meta", PredecessorSHA: strings.Repeat("c", 40),
		Selection:     Selection{ArtifactID: 1, RunID: 2, RunAttempt: 1, ReceiptDigest: receiptDigest},
		PayloadDigest: payloadDigest(raw), Receipt: raw,
	}
}
