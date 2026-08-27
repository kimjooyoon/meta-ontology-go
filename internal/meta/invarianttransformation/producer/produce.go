package producer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

// Build emits a data-only receipt. It never writes the source tree and never
// sets mutation authority; an approved artifact is a separately recorded
// effect, not a repository write.
func Build(source []byte, headSHA, caseID string) (model.Receipt, error) {
	contract := model.CanonicalContract()
	var spec model.CaseSpec
	for _, candidate := range contract.Cases {
		if candidate.ID == caseID {
			spec = candidate
			break
		}
	}
	if spec.ID == "" {
		return model.Receipt{}, fmt.Errorf("unknown invariant transformation case %q", caseID)
	}
	if !model.ValidHead(headSHA) {
		return model.Receipt{}, fmt.Errorf("invalid head sha %q", headSHA)
	}

	statuses := map[string]string{
		"precondition":       model.StatusDischarged,
		"transformation":     model.StatusDischarged,
		"postcondition":      model.StatusDischarged,
		"regression-witness": model.StatusDischarged,
	}
	reasons := map[string]string{
		"precondition":       "EXACT_SOURCE_SNAPSHOT",
		"transformation":     "TRANSFORMATION_OBSERVED",
		"postcondition":      "SEMANTIC_POSTCONDITION_PRESERVED",
		"regression-witness": "REGRESSION_REPLAY_MATCHED",
	}
	if spec.Kind == "VIOLATION" {
		statuses["postcondition"] = model.StatusRefuted
		statuses["regression-witness"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}
	if spec.Kind == "EVIDENCE_MISSING" {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = "REGRESSION_WITNESS_MISSING"
	}

	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.DigestBytes([]byte("semantic-before\x00" + headSHA))
	semanticAfter := semanticBefore
	if spec.Kind == "VIOLATION" {
		semanticAfter = model.DigestBytes([]byte("semantic-after-refuted\x00" + headSHA))
	}
	transformationDigest := model.Digest([]string{sourceDigest, semanticBefore, semanticAfter, spec.Kind})
	regressionDigest := ""
	if spec.Kind != "EVIDENCE_MISSING" {
		regressionDigest = model.Digest([]string{semanticBefore, semanticAfter, "replay-1"})
	}
	evidence := model.TransformationEvidence{
		SourceDigest: sourceDigest, CandidateDigest: transformationDigest,
		SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
		RegressionWitnessPresent: spec.Kind != "EVIDENCE_MISSING", ReplayCount: 0,
	}
	if evidence.RegressionWitnessPresent {
		evidence.ReplayBeforeDigest = semanticBefore
		evidence.ReplayAfterDigest = semanticAfter
		evidence.ReplayCount = 1
	}

	claims := make([]model.Claim, 0, len(contract.Values))
	values := make([]model.MetaValue, 0, len(contract.Values))
	for _, valueSpec := range contract.Values {
		evidence := evidenceFor(valueSpec.ID, sourceDigest, transformationDigest, semanticBefore, semanticAfter, regressionDigest)
		claim := model.Claim{ID: valueSpec.ID, Status: statuses[valueSpec.ID], Reason: reasons[valueSpec.ID],
			Coordinate:      model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]},
			EvidenceDigests: evidenceDigests(evidence), Transitions: []model.Transition{{From: model.StatusOpen, To: statuses[valueSpec.ID], Coordinate: model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]}}}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: valueSpec.ID, Kind: valueSpec.Kind, Value: statuses[valueSpec.ID], EvidenceDigest: evidence,
			Producer: valueSpec.Producer, Consumer: valueSpec.Consumer, MetaOperation: valueSpec.MetaOperation, ProofChoice: valueSpec.ProofChoice,
			Coordinate: claim.Coordinate})
	}

	decision, resolution, reason := deriveDecision(claims)
	receipt := model.Receipt{Schema: model.ReceiptSchema, CaseID: spec.ID, CaseKind: spec.Kind, HeadSHA: headSHA,
		SourcePath: model.SourcePath, SourceDigest: sourceDigest, ContractDigest: model.Digest(contract), Producer: model.ProducerID,
		Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp, ProofChoice: model.ProofRegression, Values: values, Claims: claims, Evidence: evidence,
		Decision: decision, Resolution: resolution, Reason: reason, Effects: []model.Effect{}, RepositoryWrites: 0, MutationAuthority: false}
	if spec.Kind == "APPROVED_ARTIFACT" {
		receipt.Effects = append(receipt.Effects, model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved",
			ArtifactDigest: transformationDigest, Producer: model.ProducerID, Consumer: model.ConsumerID,
			MetaOperation: "record-approved-artifact-effect", RepositoryWrites: 0, MutationAuthority: false})
	}
	return model.SealReceipt(receipt), nil
}

func evidenceFor(id, sourceDigest, transformationDigest, before, after, regression string) string {
	switch id {
	case "precondition":
		return sourceDigest
	case "transformation":
		return transformationDigest
	case "postcondition":
		return model.Digest([]string{before, after})
	case "regression-witness":
		return regression
	default:
		return ""
	}
}

func evidenceDigests(digest string) []string {
	if digest == "" {
		return []string{}
	}
	return []string{digest}
}

func deriveDecision(claims []model.Claim) (string, string, string) {
	for _, claim := range claims {
		if claim.Status == model.StatusRefuted {
			return model.DecisionRefuted, model.ResolutionInvariant, claim.Reason
		}
	}
	for _, claim := range claims {
		if claim.Status == model.StatusOpen {
			return model.DecisionBlocked, model.ResolutionLower, claim.Reason
		}
	}
	return model.DecisionAllowed, model.ResolutionExact, "ALL_INVARIANTS_DISCHARGED"
}
