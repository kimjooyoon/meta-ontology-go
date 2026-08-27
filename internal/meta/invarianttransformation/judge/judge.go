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

// Judge is intentionally implemented without importing producer. It parses
// the checked-in Gooo source independently, executes both baseline and replay
// recipes, and derives authority from independently recomputed evidence.
func Judge(receipt model.Receipt, source []byte) model.Judgment {
	if receipt.Schema != model.ReceiptSchema || !model.ValidHead(receipt.HeadSHA) ||
		receipt.SourcePath != model.SourcePath || !model.ValidDigest(receipt.SourceDigest) ||
		receipt.SourceDigest != model.DigestBytes(source) ||
		receipt.ContractDigest != model.Digest(model.CanonicalContract()) ||
		receipt.AuthorityScope != model.AuthorityScope {
		return invalid("RECEIPT_IDENTITY_INVALID")
	}
	if receipt.Digest == "" || receipt.Digest != model.SealReceipt(receipt).Digest {
		return invalid("RECEIPT_DIGEST_INVALID")
	}
	contract := model.CanonicalContract()
	if receipt.Producer != model.ProducerID || receipt.Consumer != model.ConsumerID ||
		receipt.MetaOperation != model.AuthorityOp || receipt.ProofChoice != model.ProofRegression ||
		len(receipt.Claims) != len(contract.Values) || len(receipt.Values) != len(contract.Values) {
		return invalid("RECEIPT_CONTRACT_BINDING_INVALID")
	}
	spec, ok := caseSpec(contract, receipt.CaseID, receipt.CaseKind)
	if !ok {
		return invalid("RECEIPT_CONTRACT_BINDING_INVALID")
	}
	semantics, err := parseSourceSemantics(source, spec)
	if err != nil || !validTransformationEvidence(receipt, semantics) {
		return invalid("TRANSFORMATION_EVIDENCE_INVALID")
	}

	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects)}
	for index, valueSpec := range contract.Values {
		claim := receipt.Claims[index]
		value := receipt.Values[index]
		if claim.ID != valueSpec.ID || value.ID != valueSpec.ID || value.Kind != valueSpec.Kind || value.Value != claim.Status ||
			value.Producer != valueSpec.Producer || value.Consumer != valueSpec.Consumer || value.MetaOperation != valueSpec.MetaOperation ||
			value.ProofChoice != valueSpec.ProofChoice || value.Coordinate.Stage != valueSpec.Coordinate.Stage ||
			value.Coordinate.Step != valueSpec.Coordinate.Step || value.Coordinate.Reason != claim.Reason ||
			claim.Coordinate.Stage != valueSpec.Coordinate.Stage || claim.Coordinate.Step != valueSpec.Coordinate.Step ||
			claim.Coordinate.Reason != claim.Reason || !validStatus(claim.Status) || len(claim.Transitions) != 1 {
			return invalid("CLAIM_BINDING_INVALID")
		}
		transition := claim.Transitions[0]
		if transition.From != model.StatusOpen || transition.To != claim.Status || transition.Coordinate != claim.Coordinate {
			return invalid("CLAIM_TRANSITION_INVALID")
		}
		if claim.Status == model.StatusOpen && len(claim.EvidenceDigests) != 0 {
			return invalid("OPEN_CLAIM_HAS_EVIDENCE")
		}
		if claim.Status != model.StatusOpen && !allValidDigests(claim.EvidenceDigests) {
			return invalid("CLAIM_EVIDENCE_INVALID")
		}
		if value.EvidenceDigest != firstEvidence(claim.EvidenceDigests) ||
			firstEvidence(claim.EvidenceDigests) != expectedEvidence(receipt, valueSpec.ID) ||
			claim.Status != expectedClaimStatus(receipt.Evidence, valueSpec.ID) {
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
	if receipt.RepositoryWrites != 0 || receipt.MutationAuthority {
		return invalid("WRITE_BOUNDARY_ESCALATED")
	}
	for _, effect := range receipt.Effects {
		if !validApprovedEffect(effect, semantics) {
			return invalid("UNAPPROVED_EFFECT")
		}
	}
	if !semantics.ApprovedArtifact && len(receipt.Effects) != 0 {
		return invalid("EFFECT_ON_NON_APPROVED_CASE")
	}
	if semantics.ApprovedArtifact && len(receipt.Effects) != 1 {
		return invalid("APPROVED_ARTIFACT_EFFECT_MISSING")
	}
	if len(receipt.Effects) != spec.ExpectedEffects {
		return invalid("EFFECT_CONTRACT_MISMATCH")
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
	return model.Judgment{Decision: model.DecisionRefuted, Resolution: model.ResolutionInvariant,
		Reason: reason, Status: model.StatusRefuted, Independent: false}
}

func caseSpec(contract model.Contract, id, kind string) (model.CaseSpec, bool) {
	for _, spec := range contract.Cases {
		if spec.ID == id && spec.Kind == kind {
			return spec, true
		}
	}
	return model.CaseSpec{}, false
}

func validTransformationEvidence(receipt model.Receipt, semantics sourceSemantics) bool {
	evidence := receipt.Evidence
	if evidence.SourceDigest != receipt.SourceDigest || !model.ValidDigest(evidence.SourceDigest) ||
		evidence.InputValue != semantics.Input || evidence.CandidateOperation != semantics.CandidateOperation ||
		evidence.CandidateResult != semantics.CandidateResult || evidence.ExpectedValue != semantics.Expected ||
		evidence.Invariant != semantics.Invariant || evidence.ReplayRecipe != semantics.ReplayRecipe ||
		!model.ValidDigest(evidence.CandidateDigest) || !model.ValidDigest(evidence.SemanticBeforeDigest) ||
		!model.ValidDigest(evidence.SemanticAfterDigest) || !model.ValidDigest(evidence.ExpectedSemanticDigest) ||
		evidence.CandidateDigest != model.CandidateDigest(semantics.CandidateOperation, semantics.Input, semantics.CandidateResult) ||
		evidence.SemanticBeforeDigest != model.SemanticDigest(semantics.Input) ||
		evidence.SemanticAfterDigest != model.SemanticDigest(semantics.CandidateResult) ||
		evidence.ExpectedSemanticDigest != model.SemanticDigest(semantics.Expected) ||
		evidence.BaselineInputValue != semantics.Input || evidence.BaselineOperation != semantics.CandidateOperation ||
		evidence.BaselineOutput != semantics.CandidateResult || evidence.BaselineDigest != evidence.CandidateDigest {
		return false
	}

	replayOutput, replayErr := executeReplay(semantics.ReplayRecipe, semantics.Input)
	if replayErr != nil {
		reason := replayExecutionReason
		if semantics.ReplayRecipe == "unavailable" {
			reason = replayUnavailableReason
		}
		return evidence.ReplayCount == 1 && !evidence.RegressionWitnessPresent && evidence.ReplayInputValue == 0 &&
			evidence.ReplayOperation == "" && evidence.ReplayOutput == 0 && evidence.ReplayDigest == "" &&
			evidence.ReplaySemanticDigest == "" && evidence.ReplayEvidenceDigest == "" &&
			evidence.ReplayFailureStage == "REGRESSION" && evidence.ReplayFailureStep == "execute-replay" &&
			evidence.ReplayFailureReason == reason
	}

	replayDigest := model.CandidateDigest(semantics.ReplayRecipe, semantics.Input, replayOutput)
	return evidence.ReplayCount == 2 && evidence.ReplayInputValue == semantics.Input &&
		evidence.ReplayOperation == semantics.ReplayRecipe && evidence.ReplayOutput == replayOutput &&
		evidence.ReplayDigest == replayDigest && evidence.ReplaySemanticDigest == model.SemanticDigest(replayOutput) &&
		evidence.ReplayEvidenceDigest == model.ReplayDigest(evidence.BaselineDigest, replayDigest) &&
		evidence.RegressionWitnessPresent == (evidence.BaselineDigest == replayDigest && evidence.SemanticAfterDigest == evidence.ReplaySemanticDigest) &&
		evidence.ReplayFailureStage == "" && evidence.ReplayFailureStep == "" && evidence.ReplayFailureReason == ""
}

func executeReplay(recipe string, input int64) (int64, error) {
	if recipe == "unavailable" {
		return 0, fmt.Errorf("%s", replayUnavailableReason)
	}
	return evaluateAdd(recipe, input)
}

func validApprovedEffect(effect model.Effect, semantics sourceSemantics) bool {
	if effect.Kind != model.EffectApproved || effect.ArtifactID == "" || !model.ValidDigest(effect.ArtifactDigest) ||
		effect.Producer != model.ProducerID || effect.Consumer != model.ConsumerID || effect.MetaOperation != "record-approved-artifact-effect" ||
		effect.RepositoryWrites != 0 || effect.MutationAuthority || effect.ArtifactPath != approvedArtifactPath() || effect.ArtifactSize <= 0 {
		return false
	}
	data, err := os.ReadFile(effect.ArtifactPath)
	if err != nil || len(data) != effect.ArtifactSize || model.DigestBytes(data) != effect.ArtifactDigest {
		return false
	}
	expected := []byte(fmt.Sprintf("gooo approved artifact\ncase=%s\ninput=%d\noperation=%s\noutput=%d\n",
		semantics.CaseID, semantics.Input, semantics.CandidateOperation, semantics.CandidateResult))
	return string(data) == string(expected)
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
		if evidence.ReplayCount != 2 {
			return ""
		}
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

func validStatus(status string) bool {
	return status == model.StatusOpen || status == model.StatusDischarged || status == model.StatusRefuted
}

func allValidDigests(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if !model.ValidDigest(value) {
			return false
		}
	}
	return true
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

func approvedArtifactPath() string {
	root := os.Getenv("RUNNER_TEMP")
	if root == "" {
		root = os.TempDir()
	}
	return filepath.Join(root, "gooo-invariant-transformation-approved-artifact.bin")
}
