package ambiguitybudget

import (
	"fmt"
	"reflect"
	"strings"
)

func buildInterventions(contract Contract, raw []byte, base SourceObservation, budget IntegerSet) []InterventionReceipt {
	results := make([]InterventionReceipt, 0, len(contract.Interventions))
	for _, spec := range contract.Interventions {
		mutated, err := interventionSource(raw, spec)
		if err != nil {
			results = append(results, failedIntervention(spec, err.Error()))
			continue
		}
		after, err := observeSource(base.Path, mutated)
		if err != nil {
			results = append(results, failedIntervention(spec, err.Error()))
			continue
		}
		beforeProgram, beforeOK := findCase(base, spec.TargetActivity)
		afterProgram, afterOK := findCase(after, spec.TargetActivity)
		if !beforeOK || !afterOK {
			results = append(results, failedIntervention(spec, "target case is not an observed computes declaration"))
			continue
		}
		before := caseReceipt(base.Digest, beforeProgram, budget)
		afterResult := caseReceipt(after.Digest, afterProgram, budget)
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
			SemanticDigestAfter: after.SemanticDigest, CountsBefore: before.Counts, CountsAfter: afterResult.Counts,
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
	return mutateActivityProgram(raw, spec.TargetActivity, func(program computesProgram) (string, error) {
		if program.Kind != "CASE" {
			return "", fmt.Errorf("semantic intervention target is not a case")
		}
		program.Counts.InterpretationCandidates++
		return formatComputesProgram(program), nil
	})
}

func mutateActivityProgram(raw []byte, target string, transform func(computesProgram) (string, error)) ([]byte, error) {
	source := string(raw)
	lines := strings.SplitAfter(source, "\n")
	for index, line := range lines {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 || fields[0] != "activity" {
			continue
		}
		name, _, _ := strings.Cut(fields[1], "(")
		if name != target {
			continue
		}
		marker := ` computes "`
		start := strings.Index(line, marker)
		if start < 0 {
			return nil, fmt.Errorf("target activity %q has no computes program", target)
		}
		start += len(marker)
		end := strings.LastIndex(line, `"`)
		if end <= start {
			return nil, fmt.Errorf("target activity %q has malformed computes program", target)
		}
		program, err := parseComputesProgram(target, line[start:end])
		if err != nil {
			return nil, err
		}
		replacement, err := transform(program)
		if err != nil {
			return nil, err
		}
		lines[index] = line[:start] + replacement + line[end:]
		return []byte(strings.Join(lines, "")), nil
	}
	return nil, fmt.Errorf("target activity %q not found", target)
}

func interventionSatisfied(kind string, before, after CaseReceipt, base, next SourceObservation) bool {
	if kind == "SEMANTIC" {
		return base.Digest != next.Digest && base.SemanticDigest != next.SemanticDigest && before.Counts != after.Counts &&
			before.Decision == "PASS" && before.Resolution == "EXACT" && before.Claim.From == "OPEN" && before.Claim.To == "DISCHARGED" &&
			after.Decision == "FAIL_CLOSED" && after.Resolution == "LOWER_RESOLUTION" && after.Reason == "AMBIGUITY_BUDGET_EXCEEDED" &&
			after.Claim.From == "OPEN" && after.Claim.To == "REFUTED"
	}
	return base.Digest != next.Digest && base.SemanticDigest == next.SemanticDigest && before.Counts == after.Counts &&
		before.Class == after.Class && before.InputState == after.InputState && before.Decision == after.Decision &&
		before.Resolution == after.Resolution && before.Reason == after.Reason && reflect.DeepEqual(before.UnobservedDimensions, after.UnobservedDimensions) &&
		before.Claim == after.Claim
}

func failedIntervention(spec InterventionContract, reason string) InterventionReceipt {
	return InterventionReceipt{ID: spec.ID, Kind: spec.Kind, TargetActivity: spec.TargetActivity, ReasonAfter: reason}
}
