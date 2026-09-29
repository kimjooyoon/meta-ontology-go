package lsp

import (
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/valueexecution"
)

// GeneratedReplayEvidencePart01 binds a runtime-plan replay observation to
// the current source and typed-plan identities. Digests are observations,
// never execution or adoption authority.
type GeneratedReplayEvidencePart01 struct {
	SourceDigest             string `json:"source_digest"`
	SemanticDigest           string `json:"semantic_digest"`
	TypedPlanDigest          string `json:"typed_plan_digest"`
	RuntimePlanDigest        string `json:"runtime_plan_digest"`
	GeneratedArtifactDigest  string `json:"generated_artifact_digest"`
	ReverseObservationDigest string `json:"reverse_observation_digest"`
}

// StageDigests returns the stable source-to-replay order. Invalid or absent
// digests become empty stages so the first unresolved boundary is retained.
func (evidence GeneratedReplayEvidencePart01) StageDigests() []string {
	values := evidence.rawStageDigestsPart01()
	for index, value := range values {
		values[index] = canonicalLSPProvenanceDigestPart01(value)
	}
	return values
}

func (evidence GeneratedReplayEvidencePart01) rawStageDigestsPart01() []string {
	return []string{
		evidence.SourceDigest,
		evidence.SemanticDigest,
		evidence.TypedPlanDigest,
		evidence.RuntimePlanDigest,
		evidence.GeneratedArtifactDigest,
		evidence.ReverseObservationDigest,
	}
}

// ObserveGeneratedReplayEvidencePrefixPart01 observes the ordered replay
// stages without inferring or authorizing missing evidence.
func ObserveGeneratedReplayEvidencePrefixPart01(
	evidence GeneratedReplayEvidencePart01,
) ExecutionEvidencePrefixObservation {
	return ObserveExecutionEvidencePrefix(evidence.StageDigests())
}

// ObserveGeneratedReplayEvidenceReceiptClosurePart01 closes replay evidence
// only when the execution receipt names the same runtime plan.
func ObserveGeneratedReplayEvidenceReceiptClosurePart01(
	evidence GeneratedReplayEvidencePart01,
	receipt valueexecution.ExecutionOriginReceipt,
) ExecutionEvidenceReceiptClosureObservation {
	prefix := ObserveGeneratedReplayEvidencePrefixPart01(evidence)
	observation := ObserveExecutionEvidenceReceiptClosure(prefix, receipt)
	for index, value := range evidence.rawStageDigestsPart01() {
		if strings.TrimSpace(value) == "" {
			return generatedReplayClosureAtStagePart01(
				observation,
				index,
				"GENERATED_REPLAY_"+generatedReplayStageNamePart01(index)+"_DIGEST_MISSING",
			)
		}
		if !knownLSPProvenanceDigest(value) {
			return generatedReplayClosureAtStagePart01(
				observation,
				index,
				"GENERATED_REPLAY_"+generatedReplayStageNamePart01(index)+"_DIGEST_INVALID",
			)
		}
	}
	if (prefix.MissingStageIndex < 0 || prefix.MissingStageIndex > 3) &&
		receipt.Validate() && receipt.Status == valueexecution.ExecutionOriginStatusBound &&
		receipt.Phase == valueexecution.ExecutionPhaseCompleted &&
		canonicalLSPProvenanceDigestPart01(receipt.RuntimePlanDigest) !=
			canonicalLSPProvenanceDigestPart01(evidence.RuntimePlanDigest) {
		observation = generatedReplayClosureAtStagePart01(
			observation,
			3,
			"EXECUTION_RUNTIME_PLAN_DIGEST_MISMATCH",
		)
	}
	return observation
}

func ValidateGeneratedReplayEvidenceReceiptClosurePart01(
	observation ExecutionEvidenceReceiptClosureObservation,
	evidence GeneratedReplayEvidencePart01,
	receipt valueexecution.ExecutionOriginReceipt,
) bool {
	expected := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidence, receipt)
	return observation.Validate() &&
		observation.Schema == expected.Schema &&
		observation.EvidencePrefixDigest == expected.EvidencePrefixDigest &&
		observation.ReceiptDigest == expected.ReceiptDigest &&
		observation.MissingStageIndex == expected.MissingStageIndex &&
		observation.Status == expected.Status &&
		observation.Reason == expected.Reason &&
		observation.NonAuthorizing == expected.NonAuthorizing &&
		observation.Digest == expected.Digest
}

// executionPlanGeneratedReplayDigestsPart01 accepts generation evidence only
// when it is bound to this document and its current typed plan. The generated
// artifact can remain an observed stage while replay is UNKNOWN; the reverse
// stage is included only after a valid completed receipt closes the prefix.
func executionPlanGeneratedReplayDigestsPart01(
	params ExecutionPlanProvenanceParamsPart01,
	sourceDigest,
	semanticDigest,
	typedPlanDigest string,
) (string, string, ExecutionEvidenceReceiptClosureObservation) {
	evidence := params.GeneratedReplayEvidence
	evidenceValue := GeneratedReplayEvidencePart01{}
	if evidence != nil {
		evidenceValue = *evidence
	}
	receipt := valueexecution.ExecutionOriginReceipt{}
	if params.ExecutionOriginReceipt != nil {
		receipt = *params.ExecutionOriginReceipt
	}
	closure := ObserveGeneratedReplayEvidenceReceiptClosurePart01(evidenceValue, receipt)
	if evidence == nil {
		return "", "", closure
	}
	for index, identity := range []struct {
		name     string
		observed string
		current  string
	}{
		{name: "SOURCE", observed: evidence.SourceDigest, current: sourceDigest},
		{name: "SEMANTIC", observed: evidence.SemanticDigest, current: semanticDigest},
		{name: "TYPED_PLAN", observed: evidence.TypedPlanDigest, current: typedPlanDigest},
	} {
		reason := generatedReplayIdentityFailureReasonPart01(
			identity.observed,
			identity.current,
			identity.name,
		)
		if reason != "" {
			closure = generatedReplayClosureAtStagePart01(closure, index, reason)
			return "", "", closure
		}
	}
	generatedArtifactDigest := canonicalLSPProvenanceDigestPart01(evidence.GeneratedArtifactDigest)
	if closure.MissingStageIndex >= 0 && closure.MissingStageIndex <= 4 {
		return "", "", closure
	}
	if generatedArtifactDigest == "" {
		return "", "", generatedReplayClosureAtStagePart01(closure, 4, "GENERATED_REPLAY_ARTIFACT_DIGEST_INVALID")
	}
	if closure.Status == ExecutionEvidenceReceiptClosureComplete {
		return generatedArtifactDigest, closure.Digest, closure
	}
	return generatedArtifactDigest, "", closure
}

func generatedReplayIdentityFailureReasonPart01(observed, current, stage string) string {
	switch {
	case strings.TrimSpace(observed) == "":
		return "GENERATED_REPLAY_" + stage + "_DIGEST_MISSING"
	case !knownLSPProvenanceDigest(observed):
		return "GENERATED_REPLAY_" + stage + "_DIGEST_INVALID"
	case canonicalLSPProvenanceDigestPart01(current) == "":
		return "GENERATED_REPLAY_CURRENT_" + stage + "_DIGEST_UNAVAILABLE"
	case !sameLSPProvenanceDigestPart01(observed, current):
		return "GENERATED_REPLAY_" + stage + "_DIGEST_MISMATCH"
	default:
		return ""
	}
}

func generatedReplayStageNamePart01(index int) string {
	stages := [...]string{"SOURCE", "SEMANTIC", "TYPED_PLAN", "RUNTIME_PLAN", "ARTIFACT", "REVERSE_OBSERVATION"}
	if index < 0 || index >= len(stages) {
		return "UNKNOWN_STAGE"
	}
	return stages[index]
}

func generatedReplayClosureAtStagePart01(
	observation ExecutionEvidenceReceiptClosureObservation,
	stageIndex int,
	reason string,
) ExecutionEvidenceReceiptClosureObservation {
	observation.Status = ExecutionEvidenceReceiptClosureUnknown
	observation.MissingStageIndex = stageIndex
	observation.Reason = reason
	observation.Digest = observation.ComputedDigest()
	return observation
}

func canonicalLSPProvenanceDigestPart01(value string) string {
	value = strings.TrimSpace(value)
	if !knownLSPProvenanceDigest(value) {
		return ""
	}
	return strings.TrimPrefix(value, "sha256:")
}

func sameLSPProvenanceDigestPart01(left, right string) bool {
	left = canonicalLSPProvenanceDigestPart01(left)
	right = canonicalLSPProvenanceDigestPart01(right)
	return left != "" && left == right
}
