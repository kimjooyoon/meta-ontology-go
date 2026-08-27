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
	fixture, err := parseSourceFixture(source, spec)
	if err != nil {
		return model.Receipt{}, err
	}

	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.SemanticDigest(fixture.Input)
	semanticAfter := model.SemanticDigest(fixture.CandidateResult)
	expectedSemantic := model.SemanticDigest(fixture.Expected)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	postconditionDigest := model.PostconditionDigest(semanticBefore, semanticAfter, expectedSemantic)
	regressionDigest := ""
	if fixture.RegressionAvailable {
		regressionDigest = model.ReplayDigest(semanticBefore, semanticAfter)
	}
	evidence := model.TransformationEvidence{
		SourceDigest: sourceDigest, InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation,
		CandidateResult: fixture.CandidateResult, ExpectedValue: fixture.Expected, Invariant: fixture.Invariant,
		CandidateDigest: candidateDigest, SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter,
		ExpectedSemanticDigest: expectedSemantic, RegressionWitnessPresent: fixture.RegressionAvailable, ReplayCount: 0,
	}
	if fixture.RegressionAvailable {
		evidence.ReplayBeforeDigest = semanticBefore
		evidence.ReplayAfterDigest = semanticAfter
		evidence.ReplayCount = 1
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
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if !fixture.RegressionAvailable {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = "REGRESSION_WITNESS_MISSING"
	} else if fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}

	claims := make([]model.Claim, 0, len(contract.Values))
	values := make([]model.MetaValue, 0, len(contract.Values))
	for _, valueSpec := range contract.Values {
		evidence := evidenceFor(valueSpec.ID, sourceDigest, candidateDigest, postconditionDigest, regressionDigest)
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
	if fixture.ApprovedArtifact {
		receipt.Effects = append(receipt.Effects, model.Effect{Kind: model.EffectApproved, ArtifactID: "gooo://invariant-transformation/artifact/approved",
			ArtifactDigest: candidateDigest, Producer: model.ProducerID, Consumer: model.ConsumerID,
			MetaOperation: "record-approved-artifact-effect", RepositoryWrites: 0, MutationAuthority: false})
	}
	return model.SealReceipt(receipt), nil
}

func evidenceFor(id, sourceDigest, candidate, postcondition, regression string) string {
	switch id {
	case "precondition":
		return sourceDigest
	case "transformation":
		return candidate
	case "postcondition":
		return postcondition
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
