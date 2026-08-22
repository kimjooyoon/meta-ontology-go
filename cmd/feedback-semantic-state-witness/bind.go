package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackstate"
)

func bindSemanticInput(input feedbackpredecessor.Input, receipt predecessorReceipt) (feedbackstate.Input, error) {
	if receipt.Schema != predecessorReceiptSchema || receipt.Report.Schema != feedbackpredecessor.Schema ||
		!receipt.ReplayVerified || receipt.ReplayReportDigest != receipt.Report.ReportDigest ||
		receipt.ReceiptDigest == "" {
		return feedbackstate.Input{}, fmt.Errorf("predecessor selection receipt is unverified")
	}
	if receipt.Report.Repository != input.Repository || receipt.Report.PredecessorSHA != input.PredecessorSHA ||
		receipt.Report.Decision != feedbackpredecessor.DecisionSelected || receipt.Report.Selected == nil {
		return feedbackstate.Input{}, fmt.Errorf("predecessor selection identity is unbound")
	}
	selected := receipt.Report.Selected
	matches := make([]feedbackpredecessor.Candidate, 0, 1)
	for _, candidate := range input.Candidates {
		if candidate.ArtifactID == selected.ArtifactID && candidate.RunID == selected.RunID &&
			candidate.RunAttempt == selected.RunAttempt && candidate.ReceiptDigest == selected.ReceiptDigest {
			matches = append(matches, candidate)
		}
	}
	if len(matches) != 1 {
		return feedbackstate.Input{}, fmt.Errorf("selected predecessor payload count = %d", len(matches))
	}
	candidate := matches[0]
	payload, err := base64.StdEncoding.Strict().DecodeString(candidate.ReceiptPayload)
	if err != nil {
		return feedbackstate.Input{}, fmt.Errorf("decode selected predecessor payload: %w", err)
	}
	return feedbackstate.Input{
		Repository: input.Repository, PredecessorSHA: input.PredecessorSHA,
		Selection: feedbackstate.Selection{ArtifactID: selected.ArtifactID, RunID: selected.RunID,
			RunAttempt: selected.RunAttempt, ReceiptDigest: selected.ReceiptDigest},
		PayloadDigest: candidate.PayloadDigest, Receipt: json.RawMessage(payload),
		RepositoryWrites: receipt.RepositoryWrites + receipt.Report.Summary.RepositoryWrites,
	}, nil
}
