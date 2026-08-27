package judge

import (
	"fmt"
	"reflect"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/invarianttransformation/model"
)

func Indicators(summary model.Summary) []model.Indicator {
	return []model.Indicator{
		indicator("gooo.metric.invariant-transformation.case-coverage-bps.v1", model.ProofFoundation, "freeze-invariant-denominator", summary.CasesSatisfied*10_000/summary.CasesTotal, 10_000, "="),
		indicator("gooo.metric.invariant-transformation.authorized-cases.v1", model.ProofCoherence, model.AuthorityOp, summary.AuthorizedCases, 2, "="),
		indicator("gooo.metric.invariant-transformation.refuted-counterexamples.v1", model.ProofRegression, "refute-semantic-postcondition", summary.RefutedCases, 1, "="),
		indicator("gooo.metric.invariant-transformation.open-evidence-gaps.v1", model.ProofRegression, "preserve-missing-regression-witness", summary.OpenCases, 1, "="),
		indicator("gooo.metric.invariant-transformation.claim-transition-events.v1", model.ProofCoherence, "record-claim-transitions", summary.TransitionEvents, 16, "="),
		indicator("gooo.metric.invariant-transformation.approved-artifact-effects.v1", model.ProofCoherence, "record-approved-artifact-effect", summary.ApprovedArtifactEffects, 1, "="),
		indicator("gooo.metric.invariant-transformation.repository-writes.guardrail.v1", model.ProofFoundation, "deny-repository-writes", summary.RepositoryWrites, 0, "="),
		indicator("gooo.metric.invariant-transformation.mutation-authority.guardrail.v1", model.ProofFoundation, "deny-mutation-authority", summary.MutationAuthority, 0, "="),
	}
}

func ValidateReport(report model.Report) error {
	contract := model.CanonicalContract()
	if report.Schema != model.ReportSchema || !model.ValidHead(report.HeadSHA) || report.SourcePath != model.SourcePath ||
		!model.ValidDigest(report.SourceDigest) || report.ContractDigest != model.Digest(contract) || report.DenominatorID != model.DenominatorID ||
		report.DenominatorTotal != len(contract.Cases) || report.Decision != model.DecisionPass || report.Resolution != model.ResolutionExact ||
		report.Reason != "INVARIANT_TRANSFORMATION_SUITE_SATISFIED" || report.Digest != model.SealReport(report).Digest {
		return fmt.Errorf("invariant transformation report identity or decision mismatch")
	}
	if len(report.Cases) != len(contract.Cases) {
		return fmt.Errorf("invariant transformation denominator mismatch")
	}
	var summary model.Summary
	for index, result := range report.Cases {
		if result.Spec != contract.Cases[index] || result.Receipt.CaseID != result.Spec.ID || result.Receipt.SourceDigest != report.SourceDigest {
			return fmt.Errorf("case %s is not source-bound", result.Spec.ID)
		}
		judgment := Judge(result.Receipt)
		if !judgment.Independent || judgment != result.Judgment || !result.Satisfied ||
			judgment.Decision != result.Spec.ExpectedDecision || judgment.Resolution != result.Spec.ExpectedResolution ||
			judgment.Reason != result.Spec.ExpectedReason || judgment.Status != result.Spec.ExpectedStatus ||
			len(result.Receipt.Effects) != result.Spec.ExpectedEffects {
			return fmt.Errorf("case %s does not satisfy its fixed expectation", result.Spec.ID)
		}
		summary.CasesTotal++
		summary.CasesSatisfied++
		summary.ClaimsTotal += judgment.CheckedClaims
		summary.DischargedClaims += judgment.DischargedClaims
		summary.OpenClaims += judgment.OpenClaims
		summary.RefutedClaims += judgment.RefutedClaims
		summary.TransitionEvents += len(result.Receipt.Claims)
		summary.ApprovedArtifactEffects += len(result.Receipt.Effects)
		summary.RepositoryWrites += result.Receipt.RepositoryWrites
		if result.Receipt.MutationAuthority {
			summary.MutationAuthority++
		}
		switch judgment.Decision {
		case model.DecisionAllowed:
			summary.AuthorizedCases++
		case model.DecisionBlocked:
			summary.OpenCases++
		case model.DecisionRefuted:
			summary.RefutedCases++
		}
	}
	if summary.CasesTotal != 0 {
		summary.CoverageBPS = summary.CasesSatisfied * 10_000 / summary.CasesTotal
	}
	if !reflect.DeepEqual(report.Summary, summary) || !reflect.DeepEqual(report.Indicators, Indicators(summary)) {
		return fmt.Errorf("invariant transformation summary or indicators mismatch")
	}
	return nil
}

func indicator(id, proof, operation string, value, target int, relation string) model.Indicator {
	return model.Indicator{MetricID: id, Producer: model.ProducerID, Consumer: model.ConsumerID,
		MetaOperation: operation, ProofChoice: proof, Value: value, Target: target, Relation: relation, Satisfied: value == target}
}
