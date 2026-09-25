package lsp

import "github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"

// GeneratedReplayEvidencePart01 is the ordered, non-authorizing evidence
// boundary for a generated runtime plan replay. The receipt carries execution
// identity separately so an incomplete prefix cannot be promoted by omission.
type GeneratedReplayEvidencePart01 struct {
	SourceDigest             string \`json:"source_digest"\`
	SemanticDigest           string \`json:"semantic_digest"\`
	TypedPlanDigest          string \`json:"typed_plan_digest"\`
	RuntimePlanDigest        string \`json:"runtime_plan_digest"\`
	GeneratedArtifactDigest  string \`json:"generated_artifact_digest"\`
	ReverseObservationDigest string \`json:"reverse_observation_digest"\`
}

// StageDigests returns the fixed declaration-to-replay order used by the LSP
// evidence prefix. Empty values remain visible so the first missing stage is
// preserved by ObserveExecutionEvidencePrefix.
func (evidence GeneratedReplayEvidencePart01) StageDigests() []string {
	return []string{
		evidence.SourceDigest,
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
		evidence.RuntimePlanDigest,
		evidence.GeneratedArtifactDigest,
		evidence.ReverseObservationDigest,
	}
}

// ObserveGeneratedReplayEvidencePrefixPart01 observes generated replay stages
// without granting execution, adoption, merge, or repository-write authority.
func ObserveGeneratedReplayEvidencePrefixPart01(
	evidence GeneratedReplayEvidencePart01,
) ExecutionEvidencePrefixObservation {
	return ObserveExecutionEvidencePrefix(evidence.StageDigests())
}

// ObserveGeneratedReplayEvidenceReceiptClosurePart01 binds the generated
// replay prefix to the detached execution origin receipt. Both layers remain
// non-authorizing and preserve the first unresolved stage.
func ObserveGeneratedReplayEvidenceReceiptClosurePart01(
	evidence GeneratedReplayEvidencePart01,
	receipt valueexecution.ExecutionOriginReceipt,
) ExecutionEvidenceReceiptClosureObservation {
	prefix := ObserveGeneratedReplayEvidencePrefixPart01(evidence)
	return ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
}