package generation

import (
	"encoding/json"
	"sort"
)

type artifactProvenanceInput struct {
	BaseSHA                       string `json:"base_sha"`
	HeadSHA                       string `json:"head_sha"`
	PlanDigest                    string `json:"plan_digest"`
	ExecutionManifestDigest       string `json:"execution_manifest_digest"`
	ReceiptReportDigest           string `json:"receipt_report_digest"`
	IndicatorDecisionLedgerDigest string `json:"indicator_decision_ledger_digest"`
	IndicatorDecisionLedgerCount  int    `json:"indicator_decision_ledger_count"`
}

func finishArtifactProvenance(envelope ArtifactProvenance) ArtifactProvenance {
	sort.Slice(envelope.Indicators, func(i, j int) bool {
		return envelope.Indicators[i].ID < envelope.Indicators[j].ID
	})
	envelope.Summary = summarizeArtifactProvenance(envelope.Indicators)
	switch {
	case envelope.Summary.Fail != 0:
		envelope.Decision = ArtifactProvenanceDecisionRejected
		envelope.Reason = ArtifactProvenanceReasonMismatch
	case envelope.Summary.Unknown != 0:
		envelope.Decision = ArtifactProvenanceDecisionUnknown
		envelope.Reason = ArtifactProvenanceReasonUnproven
	default:
		envelope.Decision = ArtifactProvenanceDecisionBound
		envelope.Reason = ArtifactProvenanceReasonBound
	}
	envelope.PromotionAuthorized = false
	envelope.InputDigest = digestJSON(artifactProvenanceInput{
		BaseSHA: envelope.BaseSHA, HeadSHA: envelope.HeadSHA,
		PlanDigest:                    envelope.PlanDigest,
		ExecutionManifestDigest:       envelope.ExecutionManifestDigest,
		ReceiptReportDigest:           envelope.ReceiptReportDigest,
		IndicatorDecisionLedgerDigest: envelope.IndicatorDecisionLedgerDigest,
		IndicatorDecisionLedgerCount:  envelope.IndicatorDecisionLedgerCount,
	})
	envelope.EnvelopeDigest, envelope.ReplayDigest = "", ""
	envelope.EnvelopeDigest = digestJSON(envelope)
	envelope.ReplayDigest = digestPair(envelope.InputDigest, envelope.EnvelopeDigest)
	return envelope
}

func summarizeArtifactProvenance(
	indicators []ArtifactProvenanceIndicator,
) ArtifactProvenanceSummary {
	summary := ArtifactProvenanceSummary{}
	for _, indicator := range indicators {
		switch indicator.Verdict {
		case IndicatorVerdictPass:
			summary.Pass++
		case IndicatorVerdictFail:
			summary.Fail++
		default:
			summary.Unknown++
		}
	}
	return summary
}

func EncodeArtifactProvenance(envelope ArtifactProvenance) ([]byte, error) {
	payload, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}
