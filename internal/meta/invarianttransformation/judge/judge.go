package judge

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

// Judge is intentionally implemented without importing producer. It derives
// the authority decision from the receipt's four claims and the fixed Gooo
// contract, so the actor cannot also be its own judge.
func Judge(receipt model.Receipt) model.Judgment {
	if receipt.Schema != model.ReceiptSchema || !model.ValidHead(receipt.HeadSHA) ||
		receipt.SourcePath != model.SourcePath || !model.ValidDigest(receipt.SourceDigest) ||
		receipt.ContractDigest != model.Digest(model.CanonicalContract()) {
		return invalid("RECEIPT_IDENTITY_INVALID")
	}
	if receipt.Digest == "" || receipt.Digest != model.SealReceipt(receipt).Digest {
		return invalid("RECEIPT_DIGEST_INVALID")
	}
	contract := model.CanonicalContract()
	if !knownCase(contract, receipt.CaseID, receipt.CaseKind) || receipt.Producer != model.ProducerID ||
		receipt.Consumer != model.ConsumerID || receipt.MetaOperation != model.AuthorityOp ||
		receipt.ProofChoice != model.ProofRegression || len(receipt.Claims) != len(contract.Values) ||
		len(receipt.Values) != len(contract.Values) {
		return invalid("RECEIPT_CONTRACT_BINDING_INVALID")
	}

	judgment := model.Judgment{Independent: true, CheckedClaims: len(receipt.Claims), Effects: len(receipt.Effects)}
	for index, spec := range contract.Values {
		claim := receipt.Claims[index]
		value := receipt.Values[index]
		if claim.ID != spec.ID || value.ID != spec.ID || value.Kind != spec.Kind || value.Value != claim.Status ||
			value.Producer != spec.Producer || value.Consumer != spec.Consumer || value.MetaOperation != spec.MetaOperation ||
			value.ProofChoice != spec.ProofChoice || value.Coordinate.Stage != spec.Coordinate.Stage ||
			value.Coordinate.Step != spec.Coordinate.Step || value.Coordinate.Reason != claim.Reason ||
			claim.Coordinate.Stage != spec.Coordinate.Stage || claim.Coordinate.Step != spec.Coordinate.Step ||
			claim.Coordinate.Reason != claim.Reason || !validStatus(claim.Status) || len(claim.Transitions) != 1 {
			return invalid("CLAIM_BINDING_INVALID")
		}
		transition := claim.Transitions[0]
		if transition.From != model.StatusOpen || transition.To != claim.Status ||
			transition.Coordinate != claim.Coordinate {
			return invalid("CLAIM_TRANSITION_INVALID")
		}
		if claim.Status == model.StatusOpen && len(claim.EvidenceDigests) != 0 {
			return invalid("OPEN_CLAIM_HAS_EVIDENCE")
		}
		if claim.Status != model.StatusOpen && !allValidDigests(claim.EvidenceDigests) {
			return invalid("CLAIM_EVIDENCE_INVALID")
		}
		if value.EvidenceDigest != firstEvidence(claim.EvidenceDigests) {
			return invalid("META_VALUE_EVIDENCE_MISMATCH")
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
		if effect.Kind != model.EffectApproved || effect.ArtifactID == "" || !model.ValidDigest(effect.ArtifactDigest) ||
			effect.Producer != model.ProducerID || effect.Consumer != model.ConsumerID || effect.MetaOperation != "record-approved-artifact-effect" ||
			effect.RepositoryWrites != 0 || effect.MutationAuthority {
			return invalid("UNAPPROVED_EFFECT")
		}
	}
	if receipt.CaseKind != "APPROVED_ARTIFACT" && len(receipt.Effects) != 0 {
		return invalid("EFFECT_ON_NON_APPROVED_CASE")
	}
	if receipt.CaseKind == "APPROVED_ARTIFACT" && len(receipt.Effects) != 1 {
		return invalid("APPROVED_ARTIFACT_EFFECT_MISSING")
	}

	judgment.Decision, judgment.Resolution, judgment.Reason = derive(receipt.Claims)
	judgment.Status = statusFor(judgment.Decision)
	if receipt.Decision != judgment.Decision || receipt.Resolution != judgment.Resolution || receipt.Reason != judgment.Reason {
		return invalid("DECLARED_DECISION_MISMATCH")
	}
	return judgment
}

func ValidateReceipt(receipt model.Receipt) error {
	judgment := Judge(receipt)
	if !judgment.Independent {
		return fmt.Errorf("independent judge rejected receipt: %s", judgment.Reason)
	}
	return nil
}

func invalid(reason string) model.Judgment {
	return model.Judgment{Decision: model.DecisionRefuted, Resolution: model.ResolutionInvariant,
		Reason: reason, Status: model.StatusRefuted, Independent: false}
}

func knownCase(contract model.Contract, id, kind string) bool {
	for _, spec := range contract.Cases {
		if spec.ID == id && spec.Kind == kind {
			return true
		}
	}
	return false
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
