package bodycodegen

import (
	"context"
	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

func (e *criteriaQuestionEvaluator) rank(ctx context.Context, inputs []string) (CriteriaQuestionRound, error) {
	count := min(e.receipt.Options.MaxCandidates, 1<<len(e.plan.choices))
	r := CriteriaQuestionRound{Candidates: count, Unobserved: (1 << len(e.plan.choices)) - count}
	var possible []uint16
	for mask := range count {
		candidate, err := e.classify(ctx, uint16(mask), inputs)
		r.Evaluations = append(r.Evaluations, candidate)
		if err != nil {
			return r, err
		}
		switch candidate.Status {
		case "SATISFIED":
			r.SurvivingMasks = append(r.SurvivingMasks, uint16(mask))
			possible = append(possible, uint16(mask))
		case "UNKNOWN":
			r.UnavailableMasks = append(r.UnavailableMasks, uint16(mask))
			possible = append(possible, uint16(mask))
		default:
			r.RejectedMasks = append(r.RejectedMasks, uint16(mask))
		}
	}
	used := map[string]bool{}
	for _, input := range inputs {
		key, _ := assemblyspec.RecordValueKey(input)
		used[key] = true
	}
	best := 0
	for index, input := range e.receipt.Options.CriteriaInputs {
		key, _ := assemblyspec.RecordValueKey(input)
		probe := CriteriaQuestionProbe{Input: input, AlreadyTested: used[key]}
		if !probe.AlreadyTested {
			groups, available := map[string]int{}, 0
			for _, mask := range possible {
				outputIndex, value, err := e.output(ctx, mask, input)
				probe.OutputIndices = append(probe.OutputIndices, outputIndex)
				if err != nil {
					return r, err
				}
				if value.Error == "" {
					probe.SeparatedPairs += available - groups[value.Value]
					groups[value.Value]++
					available++
				}
			}
		}
		r.Probes = append(r.Probes, probe)
		if probe.SeparatedPairs > best {
			best = probe.SeparatedPairs
			r.RecommendedIndex = new(index)
		}
	}
	if e.receipt.Options.CriteriaStrategy != "" {
		bindCriteriaQuestionOutputEvidence(&r, e.receipt.Outputs)
		r.RecommendedIndex, r.RecommendationReason = criteriaQuestionRecommendation(r)
	}
	return r, ctx.Err()
}

func (e *criteriaQuestionEvaluator) classify(ctx context.Context, mask uint16, inputs []string) (CriteriaQuestionCandidate, error) {
	r := CriteriaQuestionCandidate{Mask: mask, Status: "SATISFIED"}
	for _, c := range recordSelectionCases(e.plan) {
		index, output, err := e.output(ctx, mask, c.Inputs)
		r.ExactCaseOutputs = append(r.ExactCaseOutputs, index)
		if err != nil {
			return r, err
		}
		if output.Error != "" {
			r.Status = "UNKNOWN"
			continue
		}
		expected, err := assemblyspec.RecordValueKey(c.Expected)
		if err != nil {
			return r, err
		}
		if output.Value != expected {
			r.Status = "UNSATISFIED"
			return r, nil
		}
	}
	for _, input := range inputs {
		index, row, err := e.evaluate(ctx, mask, input)
		r.CriteriaIndices = append(r.CriteriaIndices, index)
		if err != nil {
			return r, err
		}
		failed, unknown := false, false
		for _, check := range row.Checks {
			failed = failed || check.Status == "UNSATISFIED"
			unknown = unknown || check.Status == "UNKNOWN"
		}
		if failed {
			r.Status = "UNSATISFIED"
			return r, nil
		}
		if unknown {
			r.Status = "UNKNOWN"
		}
	}
	return r, ctx.Err()
}
