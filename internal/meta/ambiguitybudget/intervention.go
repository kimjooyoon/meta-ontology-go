package ambiguitybudget

import (
	"fmt"
	"reflect"
	"strings"
)

func buildInterventions(contract Contract, raw []byte, base SourceObservation, policy BudgetPolicy) []InterventionReceipt {
	results := make([]InterventionReceipt, 0, len(contract.Interventions))
	for _, spec := range contract.Interventions {
		mutated, err := interventionSource(raw, spec)
		if err != nil {
			results = append(results, failedIntervention(spec, err.Error()))
			continue
		}
		after, err := observeSource(base.Path, mutated, policy)
		beforeProgram, beforeOK := findCase(base, spec.TargetActivity)
		afterProgram, afterOK := findCase(after, spec.TargetActivity)
		if err != nil || !beforeOK || !afterOK {
			results = append(results, failedIntervention(spec, "target case is not an observed computes declaration"))
			continue
		}
		before := caseReceipt(base.Digest, base.SemanticDigest, beforeProgram, policy)
		afterResult := caseReceipt(after.Digest, after.SemanticDigest, afterProgram, policy)
		satisfied := interventionSatisfied(spec.Kind, before, afterResult, base, after)
		evidence := digestValue(struct {
			ID     string
			Before CaseReceipt
			After  CaseReceipt
			Base   SourceObservation
			Next   SourceObservation
		}{spec.ID, before, afterResult, base, after})
		results = append(results, InterventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity,
			SourceDigestBefore: base.Digest, SourceDigestAfter: after.Digest, SemanticDigestBefore: base.SemanticDigest,
			SemanticDigestAfter: after.SemanticDigest, ElementsBefore: before.Elements, ElementsAfter: afterResult.Elements,
			CountsBefore: before.Counts, CountsAfter: afterResult.Counts,
			UnobservedBefore: append([]string(nil), before.UnobservedDimensions...), UnobservedAfter: append([]string(nil), afterResult.UnobservedDimensions...),
			ClassBefore: before.Class, ClassAfter: afterResult.Class, InputStateBefore: before.InputState, InputStateAfter: afterResult.InputState,
			ClaimBefore: before.Claim, ClaimAfter: afterResult.Claim,
			DecisionBefore: before.Decision, ResolutionBefore: before.Resolution, ReasonBefore: before.Reason,
			DecisionAfter: afterResult.Decision, ResolutionAfter: afterResult.Resolution, ReasonAfter: afterResult.Reason,
			Satisfied: satisfied, EvidenceDigest: evidence})
	}
	return results
}

func interventionSource(raw []byte, spec InterventionContract) ([]byte, error) {
	source := string(raw)
	if spec.Kind == "NONSEMANTIC" {
		needle := "activity " + spec.TargetActivity + "("
		if !strings.Contains(source, needle) {
			return nil, fmt.Errorf("target activity %q not found", spec.TargetActivity)
		}
		return []byte(strings.Replace(source, needle, "// ambiguity-budget: comment-only intervention\n"+needle, 1)), nil
	}
	if spec.Kind != "SEMANTIC" || spec.TargetActivity != "BoundaryAmbiguity" {
		return nil, fmt.Errorf("unsupported intervention kind or target")
	}
	needle := "activity BoundaryAmbiguity(BoundaryCase) -> AmbiguityReceipt computes \"ambiguity-budget:case:boundary-ambiguity\""
	if !strings.Contains(source, needle) {
		return nil, fmt.Errorf("semantic target activity not found")
	}
	addition := "entity BoundaryCandidateC id \"gooo://ambiguity-budget/case/boundary-ambiguity/candidate/c\"\n" +
		"entity BoundaryEvidencePathC id \"gooo://ambiguity-budget/case/boundary-ambiguity/evidence-path/c\"\n" +
		"activity ObserveBoundaryCandidateC(BoundaryCase) -> BoundaryCandidateC\n" +
		"activity ObserveBoundaryEvidencePathC(BoundaryCase, BoundaryCandidateC) -> BoundaryEvidencePathC\n"
	return []byte(strings.Replace(source, needle, addition+needle, 1)), nil
}

func interventionSatisfied(kind string, before, after CaseReceipt, base, next SourceObservation) bool {
	if kind == "SEMANTIC" {
		return base.Digest != next.Digest && base.SemanticDigest != next.SemanticDigest &&
			!reflect.DeepEqual(before.Elements, after.Elements) && before.Counts != after.Counts && before.Decision == "PASS" &&
			before.Resolution == "EXACT" && before.Claim.From == "OPEN" && before.Claim.To == "DISCHARGED" &&
			after.Decision == "FAIL_CLOSED" && after.Resolution == "LOWER_RESOLUTION" && after.Reason == "AMBIGUITY_BUDGET_EXCEEDED" &&
			after.Claim.From == "OPEN" && after.Claim.To == "REFUTED" && before.Proposition == after.Proposition &&
			before.Claim.PropositionDigest == after.Claim.PropositionDigest
	}
	return base.Digest != next.Digest && base.SemanticDigest == next.SemanticDigest && reflect.DeepEqual(before.Elements, after.Elements) &&
		before.Counts == after.Counts && before.Class == after.Class && before.InputState == after.InputState &&
		before.Decision == after.Decision && before.Resolution == after.Resolution && before.Reason == after.Reason &&
		reflect.DeepEqual(before.UnobservedDimensions, after.UnobservedDimensions) && before.Proposition == after.Proposition &&
		before.PropositionDigest == after.PropositionDigest && before.Claim == after.Claim
}

func failedIntervention(spec InterventionContract, reason string) InterventionReceipt {
	return InterventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity, ReasonAfter: reason}
}
