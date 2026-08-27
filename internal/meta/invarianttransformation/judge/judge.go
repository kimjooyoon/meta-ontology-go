package judge

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

const (
	replayUnavailableReason = "REGRESSION_REPLAY_RECIPE_UNAVAILABLE"
	replayExecutionReason   = "REGRESSION_REPLAY_EXECUTION_FAILED"
	replayMismatchReason    = "REGRESSION_REPLAY_MISMATCH"
)

// Judge independently reconstructs the source recipe and validates a receipt.
// It does not import producer or intervention code and never creates files.
func Judge(receipt model.Receipt, source []byte) model.Judgment {
	if receipt.Schema != model.ReceiptSchema || !model.ValidHead(receipt.HeadSHA) || receipt.SourcePath != model.SourcePath ||
		!model.ValidDigest(receipt.SourceDigest) || receipt.SourceDigest != model.DigestBytes(source) ||
		receipt.ContractDigest != model.ValueContractDigest() || receipt.ValidatorContractDigest != model.ValidatorContractDigest() ||
		receipt.AuthorityScope != model.AuthorityScope || receipt.RepositoryNetStatusUnchanged != true ||
		receipt.RepositoryActualOrTransientWrites != model.UnknownEffectScope || receipt.RepositoryWritesObserved || receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return invalid("RECEIPT_IDENTITY_INVALID")
	}
	if receipt.Digest == "" || receipt.Digest != model.SealReceipt(receipt).Digest {
		return invalid("RECEIPT_DIGEST_INVALID")
	}
	semantics, err := parseSourceSemantics(source, receipt.CaseID)
	if err != nil || receipt.CaseKind != semantics.CaseKind || receipt.ActivityStableID != semantics.ActivityID ||
		receipt.SemanticSourceDigest != semantics.SemanticSourceDigest || receipt.Producer != model.ProducerID || receipt.Consumer != model.ConsumerID ||
		receipt.MetaOperation != model.AuthorityOp || receipt.ProofChoice != model.ProofRegression || len(receipt.Claims) != len(model.CanonicalValueSpecs()) ||
		len(receipt.Values) != len(model.CanonicalValueSpecs()) || !validPhase(receipt.Phase) {
		return invalid("RECEIPT_SOURCE_BINDING_INVALID")
	}
	if !validTransformationEvidence(receipt, semantics) {
		return invalid("TRANSFORMATION_EVIDENCE_INVALID")
	}

	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects), AuthorizationDigest: receipt.AuthorizationDigest}
	for index, valueSpec := range model.CanonicalValueSpecs() {
		claim := receipt.Claims[index]
		value := receipt.Values[index]
		if claim.ID != receipt.CaseID+"::"+valueSpec.ID || value.ID != valueSpec.ID || value.Kind != valueSpec.Kind || value.Value != claim.Status ||
			value.Producer != valueSpec.Producer || value.Consumer != valueSpec.Consumer || value.MetaOperation != valueSpec.MetaOperation ||
			value.ProofChoice != valueSpec.ProofChoice || value.VerificationCheck != valueSpec.VerificationCheck || value.Coordinate.Stage != valueSpec.Coordinate.Stage ||
			value.Coordinate.Step != valueSpec.Coordinate.Step || value.Coordinate.Reason != claim.Reason || claim.VerificationCheck != valueSpec.VerificationCheck ||
			claim.Coordinate.Stage != valueSpec.Coordinate.Stage || claim.Coordinate.Step != valueSpec.Coordinate.Step || claim.Coordinate.Reason != claim.Reason ||
			!validStatus(claim.Status) || len(claim.Transitions) != 1 || len(claim.EvidenceDigests) > 1 {
			return invalid("CLAIM_BINDING_INVALID")
		}
		transition := claim.Transitions[0]
		evidence := expectedEvidence(receipt, valueSpec.ID)
		if transition.ClaimID != claim.ID || transition.From != model.StatusOpen || transition.To != claim.Status || transition.Coordinate != claim.Coordinate ||
			transition.EvidenceDigest != evidence || transition.CurrentTransitionDigest != model.TransitionDigest(transition) ||
			claim.TargetDigest != transition.PropositionDigest || claim.PriorStateDigest != transition.PriorStateDigest {
			return invalid("CLAIM_TRANSITION_INVALID")
		}
		if claim.Status == model.StatusOpen && len(claim.EvidenceDigests) != 0 {
			return invalid("OPEN_CLAIM_HAS_EVIDENCE")
		}
		if claim.Status != model.StatusOpen && (len(claim.EvidenceDigests) != 1 || !model.ValidDigest(claim.EvidenceDigests[0]) || claim.EvidenceDigests[0] != evidence) {
			return invalid("CLAIM_EVIDENCE_INVALID")
		}
		if value.EvidenceDigest != firstEvidence(claim.EvidenceDigests) || claim.Status != expectedClaimStatus(receipt.Evidence, valueSpec.ID) {
			return invalid("CLAIM_EVIDENCE_NOT_REPLAYABLE")
		}
		switch claim.Status {
		case model.StatusDischarged:
			judgment.DischargedClaims++
		case model.StatusOpen:
			judgment.OpenClaims++
		case model.StatusRefuted:
			judgment.RefutedClaims++
		}
	}

	if semantics.EffectIntent != "approved-artifact" && len(receipt.Effects) != 0 {
		return invalid("EFFECT_ON_NON_APPROVED_CASE")
	}
	if receipt.Phase == model.ReceiptProvisional && (receipt.TempArtifactWriteAuthorized || len(receipt.Effects) != 0) {
		return invalid("PROVISIONAL_RECEIPT_HAS_EFFECT")
	}
	if receipt.Phase == model.ReceiptExecuted && (!receipt.TempArtifactWriteAuthorized || len(receipt.Effects) != 1 || !validApprovedEffect(receipt.Effects[0], receipt, semantics)) {
		return invalid("EXECUTED_EFFECT_INVALID")
	}
	if semantics.EffectIntent == "approved-artifact" && receipt.Phase == model.ReceiptExecuted && len(receipt.Effects) != 1 {
		return invalid("APPROVED_ARTIFACT_EFFECT_MISSING")
	}
	if receipt.AuthorizationDigest != model.AuthorizationDigest(receipt) {
		return invalid("AUTHORIZATION_DIGEST_INVALID")
	}

	judgment.Decision, judgment.Resolution, judgment.Reason = derive(receipt.Claims)
	judgment.Status = statusFor(judgment.Decision)
	if receipt.Decision != judgment.Decision || receipt.Resolution != judgment.Resolution || receipt.Reason != judgment.Reason {
		return invalid("DECLARED_DECISION_MISMATCH")
	}
	return judgment
}

func ValidateReceipt(receipt model.Receipt, source []byte) error {
	judgment := Judge(receipt, source)
	if !judgment.Independent {
		return fmt.Errorf("independent judge rejected receipt: %s", judgment.Reason)
	}
	return nil
}

func invalid(reason string) model.Judgment {
	return model.Judgment{Decision: model.DecisionRefuted, Resolution: model.ResolutionInvariant, Reason: reason, Status: model.StatusRefuted, Independent: false}
}

func validTransformationEvidence(receipt model.Receipt, semantics sourceSemantics) bool {
	evidence := receipt.Evidence
	if evidence.SourceDigest != receipt.SourceDigest || !model.ValidDigest(evidence.SourceDigest) || evidence.SemanticSourceDigest != semantics.SemanticSourceDigest ||
		evidence.CaseStableID != semantics.CaseID || evidence.ActivityStableID != semantics.ActivityID || evidence.OperationID != semantics.OperationID ||
		evidence.InputDomainID != semantics.DomainID || evidence.InvariantID != semantics.InvariantID || evidence.EffectIntent != semantics.EffectIntent ||
		evidence.InputValue != semantics.Input || evidence.CandidateOperation != semantics.CandidateOperation || evidence.CandidateResult != semantics.CandidateResult ||
		evidence.ExpectedValue != semantics.Expected || evidence.Invariant != semantics.Invariant || evidence.ReplayRecipe != semantics.ReplayRecipe ||
		!model.ValidDigest(evidence.SemanticSourceDigest) || !model.ValidDigest(evidence.CandidateDigest) || !model.ValidDigest(evidence.SemanticBeforeDigest) ||
		!model.ValidDigest(evidence.SemanticAfterDigest) || !model.ValidDigest(evidence.ExpectedSemanticDigest) ||
		evidence.CandidateDigest != model.CandidateDigest(semantics.CandidateOperation, semantics.Input, semantics.CandidateResult) ||
		evidence.SemanticBeforeDigest != model.SemanticDigest(semantics.Input) || evidence.SemanticAfterDigest != model.SemanticDigest(semantics.CandidateResult) ||
		evidence.ExpectedSemanticDigest != model.SemanticDigest(semantics.Expected) || evidence.BaselineInputValue != semantics.Input ||
		evidence.BaselineOperation != semantics.CandidateOperation || evidence.BaselineOutput != semantics.CandidateResult || evidence.BaselineDigest != evidence.CandidateDigest {
		return false
	}
	replayOutput, replayErr := executeReplay(semantics.ReplayRecipe, semantics.Input)
	if replayErr != nil {
		reason := replayExecutionReason
		if semantics.ReplayRecipe == "unavailable" {
			reason = replayUnavailableReason
		}
		return evidence.ReplayCount == 1 && !evidence.RegressionWitnessPresent && evidence.ReplayInputValue == 0 && evidence.ReplayOperation == "" &&
			evidence.ReplayOutput == 0 && evidence.ReplayDigest == "" && evidence.ReplaySemanticDigest == "" && evidence.ReplayEvidenceDigest == "" &&
			evidence.ReplayFailureStage == "REGRESSION" && evidence.ReplayFailureStep == "execute-replay" && evidence.ReplayFailureReason == reason
	}
	replayDigest := model.CandidateDigest(semantics.ReplayRecipe, semantics.Input, replayOutput)
	return evidence.ReplayCount == 2 && evidence.ReplayInputValue == semantics.Input && evidence.ReplayOperation == semantics.ReplayRecipe &&
		evidence.ReplayOutput == replayOutput && evidence.ReplayDigest == replayDigest && evidence.ReplaySemanticDigest == model.SemanticDigest(replayOutput) &&
		evidence.ReplayEvidenceDigest == model.ReplayDigest(evidence.BaselineDigest, replayDigest) && evidence.RegressionWitnessPresent ==
		(evidence.BaselineDigest == replayDigest && evidence.SemanticAfterDigest == evidence.ReplaySemanticDigest) && evidence.ReplayFailureStage == "" &&
		evidence.ReplayFailureStep == "" && evidence.ReplayFailureReason == ""
}

func validApprovedEffect(effect model.Effect, receipt model.Receipt, semantics sourceSemantics) bool {
	if effect.Kind != model.EffectApproved || effect.CaseID != receipt.CaseID || effect.SubjectSHA != receipt.HeadSHA || effect.Intent != semantics.EffectIntent ||
		effect.AuthorizationDigest != receipt.AuthorizationDigest || effect.Producer != receipt.Producer || effect.Executor != model.ExecutorID || effect.Consumer != receipt.Consumer ||
		effect.MetaOperation != "execute-authorized-temp-artifact" || !effect.TempArtifactWriteAuthorized || !effect.RepositoryNetStatusUnchanged ||
		effect.RepositoryActualOrTransientWrites != model.UnknownEffectScope || !model.ValidDigest(effect.ArtifactDigest) || !allowedTempPath(effect.Artifact.Path) ||
		effect.Artifact.Path != effect.ArtifactPath || effect.Artifact.ContentDigest != effect.ArtifactDigest || effect.Artifact.Size != effect.ArtifactSize ||
		effect.Artifact.CaseID != receipt.CaseID || effect.Artifact.SubjectSHA != receipt.HeadSHA || effect.Artifact.AuthorizationDigest != receipt.AuthorizationDigest ||
		effect.Artifact.Producer != receipt.Producer || effect.Artifact.Executor != model.ExecutorID || effect.Artifact.Consumer != receipt.Consumer ||
		!effect.Artifact.RepositoryNetStatusUnchanged || model.EffectExecutionDigest(effect) != effect.ExecutionReceiptDigest || effect.Artifact.EffectReceiptDigest != effect.ExecutionReceiptDigest {
		return false
	}
	data, err := os.ReadFile(effect.Artifact.Path)
	if err != nil || len(data) != effect.ArtifactSize || model.DigestBytes(data) != effect.ArtifactDigest {
		return false
	}
	expected := []byte(fmt.Sprintf("gooo bounded transformation artifact\ncase=%s\ninput=%d\noperation=%s\noutput=%d\nsource=%s\nsemantic-source=%s\nauthorization=%s\nsubject=%s\n",
		receipt.CaseID, semantics.Input, semantics.CandidateOperation, semantics.CandidateResult, receipt.SourceDigest, receipt.SemanticSourceDigest, receipt.AuthorizationDigest, receipt.HeadSHA))
	return string(data) == string(expected)
}

func allowedTempPath(path string) bool {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return false
	}
	return path != root && filepath.Dir(path) == root
}

func expectedEvidence(receipt model.Receipt, valueID string) string {
	evidence := receipt.Evidence
	switch valueID {
	case "precondition":
		return evidence.SourceDigest
	case "transformation":
		return evidence.CandidateDigest
	case "postcondition":
		return model.PostconditionDigest(evidence.SemanticBeforeDigest, evidence.SemanticAfterDigest, evidence.ExpectedSemanticDigest)
	case "regression-witness":
		return evidence.ReplayEvidenceDigest
	default:
		return ""
	}
}

func expectedClaimStatus(evidence model.TransformationEvidence, valueID string) string {
	switch valueID {
	case "postcondition":
		if evidence.SemanticAfterDigest != evidence.ExpectedSemanticDigest {
			return model.StatusRefuted
		}
	case "regression-witness":
		if evidence.ReplayCount != 2 {
			return model.StatusOpen
		}
		if !evidence.RegressionWitnessPresent || evidence.SemanticAfterDigest != evidence.ExpectedSemanticDigest {
			return model.StatusRefuted
		}
	}
	return model.StatusDischarged
}

func validPhase(value string) bool {
	return value == model.ReceiptProvisional || value == model.ReceiptExecuted
}
func validStatus(value string) bool {
	return value == model.StatusOpen || value == model.StatusDischarged || value == model.StatusRefuted
}
func firstEvidence(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func derive(claims []model.Claim) (string, string, string) {
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

func statusFor(decision string) string {
	switch decision {
	case model.DecisionAllowed:
		return model.StatusDischarged
	case model.DecisionBlocked:
		return model.StatusOpen
	default:
		return model.StatusRefuted
	}
}
