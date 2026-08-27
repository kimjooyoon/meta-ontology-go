package producer

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const (
	replayUnavailableReason = "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	replayExecutionReason   = "REGRESSION_REPLAY_EXECUTION_FAILED"
	replayMismatchReason    = "REGRESSION_REPLAY_MISMATCH"
)

// Build emits only a source-bound provisional receipt. It executes the
// source-declared candidate and replay recipe, but never creates an artifact.
// Artifact execution is a separate post-judgment operation.
func Build(source []byte, headSHA, caseID string) (model.Receipt, error) {
	if !model.ValidHead(headSHA) {
		return model.Receipt{}, fmt.Errorf("invalid head sha %q", headSHA)
	}
	fixture, err := parseSourceFixture(source, caseID)
	if err != nil {
		return model.Receipt{}, err
	}

	sourceDigest := model.DigestBytes(source)
	semanticBefore := model.SemanticDigest(fixture.Input)
	semanticAfter := model.SemanticDigest(fixture.CandidateResult)
	expectedSemantic := model.SemanticDigest(fixture.Expected)
	candidateDigest := model.CandidateDigest(fixture.CandidateOperation, fixture.Input, fixture.CandidateResult)
	replayOutput, replayErr := executeReplay(fixture.ReplayRecipe, fixture.Input)
	evidence := model.TransformationEvidence{
		SourceDigest: sourceDigest, SemanticSourceDigest: fixture.SemanticSourceDigest,
		CaseStableID: fixture.CaseID, ActivityStableID: fixture.ActivityID, OperationID: fixture.OperationID,
		InputDomainID: fixture.DomainID, InvariantID: fixture.InvariantID, EffectIntent: fixture.EffectIntent,
		InputValue: fixture.Input, CandidateOperation: fixture.CandidateOperation, CandidateResult: fixture.CandidateResult,
		ExpectedValue: fixture.Expected, Invariant: fixture.Invariant, CandidateDigest: candidateDigest,
		SemanticBeforeDigest: semanticBefore, SemanticAfterDigest: semanticAfter, ExpectedSemanticDigest: expectedSemantic,
		ReplayRecipe: fixture.ReplayRecipe, BaselineInputValue: fixture.Input, BaselineOperation: fixture.CandidateOperation,
		BaselineOutput: fixture.CandidateResult, BaselineDigest: candidateDigest, ReplayCount: 1,
	}
	if replayErr != nil {
		evidence.ReplayFailureStage = "REGRESSION"
		evidence.ReplayFailureStep = "execute-replay"
		evidence.ReplayFailureReason = replayFailureReason(fixture.ReplayRecipe, replayErr)
	} else {
		replayDigest := model.CandidateDigest(fixture.ReplayRecipe, fixture.Input, replayOutput)
		replaySemanticDigest := model.SemanticDigest(replayOutput)
		evidence.ReplayInputValue = fixture.Input
		evidence.ReplayOperation = fixture.ReplayRecipe
		evidence.ReplayOutput = replayOutput
		evidence.ReplayDigest = replayDigest
		evidence.ReplaySemanticDigest = replaySemanticDigest
		evidence.ReplayEvidenceDigest = model.ReplayDigest(candidateDigest, replayDigest)
		evidence.ReplayCount = 2
		evidence.RegressionWitnessPresent = candidateDigest == replayDigest && semanticAfter == replaySemanticDigest
	}

	postconditionDigest := model.PostconditionDigest(semanticBefore, semanticAfter, expectedSemantic)
	regressionDigest := evidence.ReplayEvidenceDigest
	statuses := map[string]string{
		"precondition": model.StatusDischarged, "transformation": model.StatusDischarged,
		"postcondition": model.StatusDischarged, "regression-witness": model.StatusDischarged,
	}
	reasons := map[string]string{
		"precondition": "EXACT_SOURCE_SNAPSHOT", "transformation": "TRANSFORMATION_OBSERVED",
		"postcondition": "SEMANTIC_POSTCONDITION_PRESERVED", "regression-witness": "REGRESSION_REPLAY_MATCHED",
	}
	if fixture.CandidateResult != fixture.Expected {
		statuses["postcondition"] = model.StatusRefuted
		reasons["postcondition"] = "SEMANTIC_POSTCONDITION_REFUTED"
	}
	if evidence.ReplayCount != 2 {
		statuses["regression-witness"] = model.StatusOpen
		reasons["regression-witness"] = evidence.ReplayFailureReason
	} else if !evidence.RegressionWitnessPresent {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = replayMismatchReason
	} else if fixture.CandidateResult != fixture.Expected {
		statuses["regression-witness"] = model.StatusRefuted
		reasons["regression-witness"] = "REGRESSION_REPLAY_REFUTED"
	}

	valueSpecs := model.CanonicalValueSpecs()
	claims := make([]model.Claim, 0, len(valueSpecs))
	values := make([]model.MetaValue, 0, len(valueSpecs))
	for _, valueSpec := range valueSpecs {
		evidenceDigest := evidenceFor(valueSpec.ID, sourceDigest, candidateDigest, postconditionDigest, regressionDigest)
		coordinate := model.Coordinate{Stage: valueSpec.Coordinate.Stage, Step: valueSpec.Coordinate.Step, Reason: reasons[valueSpec.ID]}
		claimID := caseID + "::" + valueSpec.ID
		transition := model.NewTransition(claimID, model.StatusOpen, statuses[valueSpec.ID], coordinate, evidenceDigest)
		claim := model.Claim{ID: claimID, Status: statuses[valueSpec.ID], Reason: reasons[valueSpec.ID], VerificationCheck: valueSpec.VerificationCheck,
			Coordinate: coordinate, TargetDigest: transition.PropositionDigest, PriorStateDigest: transition.PriorStateDigest,
			EvidenceDigests: evidenceDigests(evidenceDigest), Transitions: []model.Transition{transition}}
		claims = append(claims, claim)
		values = append(values, model.MetaValue{ID: valueSpec.ID, Kind: valueSpec.Kind, Value: statuses[valueSpec.ID], EvidenceDigest: evidenceDigest,
			Producer: valueSpec.Producer, Consumer: valueSpec.Consumer, MetaOperation: valueSpec.MetaOperation, ProofChoice: valueSpec.ProofChoice,
			VerificationCheck: valueSpec.VerificationCheck, Coordinate: coordinate})
	}
	decision, resolution, reason := deriveDecision(claims)
	receipt := model.Receipt{
		Schema: model.ReceiptSchema, CaseID: fixture.CaseID, CaseKind: fixture.CaseKind, ActivityStableID: fixture.ActivityID, HeadSHA: headSHA,
		SourcePath: model.SourcePath, SourceDigest: sourceDigest, SemanticSourceDigest: fixture.SemanticSourceDigest,
		ContractDigest: model.ValueContractDigest(), ValidatorContractDigest: model.ValidatorContractDigest(), Producer: model.ProducerID,
		Consumer: model.ConsumerID, MetaOperation: model.AuthorityOp, ProofChoice: model.ProofRegression, Values: values, Claims: claims,
		Evidence: evidence, Decision: decision, Resolution: resolution, Reason: reason, Phase: model.ReceiptProvisional,
		Effects: []model.Effect{}, TempArtifactWriteAuthorized: false, RepositoryNetStatusUnchanged: true,
		RepositoryActualOrTransientWrites: model.UnknownEffectScope, RepositoryWrites: 0, MutationAuthority: false,
		AuthorityScope: model.AuthorityScope,
	}
	receipt.AuthorizationDigest = model.AuthorizationDigest(receipt)
	return model.SealReceipt(receipt), nil
}

func replayFailureReason(recipe string, _ error) string {
	if recipe == "unavailable" {
		return replayUnavailableReason
	}
	return replayExecutionReason
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
