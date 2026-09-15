package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/feedbackpredecessor"
	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func validateFoundationPromotionEvidence(evidence *foundationPromotionEvidence, context contextInput) error {
	if evidence == nil {
		return fmt.Errorf("evidence is missing")
	}
	if !isFoundationPromotionContext(context) || evidence.Schema != foundationPromotionEvidenceSchema || evidence.Decision != "PASS" || evidence.HumanDecision != verify.FoundationPromotionHumanDecision || evidence.Repository != context.Repository || evidence.PRNumber != context.PRNumber || evidence.BaseRef != context.BaseRef || evidence.BaseSHA != context.BaseSHA || evidence.HeadRef != verify.FoundationPromotionHeadBranch || evidence.HeadSHA != context.HeadSHA {
		return fmt.Errorf("identity or decision is not exact")
	}
	if evidence.ArtifactID <= 0 || evidence.ArtifactName != "feedback-predecessor-"+context.HeadSHA || evidence.ArtifactSize <= 0 || !validArtifactDigest(evidence.ArtifactDigest) || evidence.ArtifactRunID != context.RunID || evidence.ArtifactAttempt != context.RunAttempt {
		return fmt.Errorf("predecessor artifact is missing or unbound")
	}
	if evidence.Input.Repository != context.Repository || evidence.Input.PredecessorSHA != verify.FoundationPromotionBaseSHA || evidence.Input.CanonicalBranch != "main" || evidence.Input.CanonicalWorkflow != "CI [push full]" || len(evidence.Input.Candidates) != 0 {
		return fmt.Errorf("predecessor input identity is not exact")
	}
	expectedFoundation := feedbackpredecessor.FoundationEvidenceForConfirmedGap()
	if !reflect.DeepEqual(evidence.Input.Foundation, &expectedFoundation) {
		return fmt.Errorf("predecessor foundation evidence is not the confirmed gap")
	}
	report, err := feedbackpredecessor.Select(evidence.Input)
	if err != nil || report.Decision != feedbackpredecessor.DecisionFoundation || report.ReportDigest == "" || !reflect.DeepEqual(report, evidence.Receipt.Report) {
		return fmt.Errorf("predecessor report is not a deterministic FOUNDATION result")
	}
	if evidence.Receipt.Schema != "gooo/meta-feedback-predecessor-receipt/v1" || evidence.Receipt.ProofChoice != feedbackpredecessor.FoundationProofChoice || !reflect.DeepEqual(evidence.Receipt.Foundation, report.Foundation) || evidence.Receipt.ReplayReportDigest != report.ReportDigest || !evidence.Receipt.ReplayVerified || evidence.Receipt.ExpectedDigest != "" || evidence.Receipt.RepositoryWrites != 0 {
		return fmt.Errorf("predecessor receipt replay envelope is not exact")
	}
	receipt := evidence.Receipt
	recordedDigest := receipt.ReceiptDigest
	receipt.ReceiptDigest = ""
	payload, err := json.Marshal(receipt)
	if err != nil || "sha256:"+digestBytes(payload) != recordedDigest {
		return fmt.Errorf("predecessor receipt digest mismatch")
	}
	if len(evidence.NonRouteJobs) != len(foundationPromotionJobNames) {
		return fmt.Errorf("non-route job evidence is incomplete")
	}
	seenIDs := make(map[int64]bool, len(evidence.NonRouteJobs))
	for index, job := range evidence.NonRouteJobs {
		if job.Name != foundationPromotionJobNames[index] || job.ID <= 0 || seenIDs[job.ID] || job.Status != "completed" || job.Conclusion != "success" || job.HeadSHA != context.HeadSHA || job.RunID != context.RunID || job.RunAttempt != context.RunAttempt {
			return fmt.Errorf("non-route job %q is not a successful exact-head result", job.Name)
		}
		seenIDs[job.ID] = true
	}
	return nil
}
