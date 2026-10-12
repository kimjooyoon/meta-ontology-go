package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
)

func isCriteriaQuestionPlan(p recordAssemblyPlan) bool {
	return p.spec != nil && p.spec.Observation != nil && len(p.spec.Observation.CriteriaInputs) > 0
}

func criteriaSelectionInputs(p recordAssemblyPlan) []string {
	if p.effectiveCriteriaInputs != nil {
		return p.effectiveCriteriaInputs
	}
	if p.spec.Criteria == nil {
		return nil
	}
	return p.spec.Criteria.Inputs
}

func criteriaInputsSHA(inputs []string) string { raw, _ := json.Marshal(inputs); return digest(raw) }

func validateCriteriaQuestionInputs(ctx context.Context, p recordAssemblyPlan) error {
	e, err := newCriteriaTupleEvaluator(ctx, p.body.base.source, p.body.activity.Name, p.body.records, nil)
	if err != nil {
		return err
	}
	for _, input := range p.spec.Observation.CriteriaInputs {
		if err := e.bind(input); err != nil {
			return fmt.Errorf("criteria question: %w", err)
		}
	}
	return ctx.Err()
}

func observeRecordCriteriaQuestions(ctx context.Context, source []byte, p *recordAssemblyPlan, receipt *RecordAssemblyReceipt) (failure error) {
	o := *p.spec.Clone().Observation
	r := &CriteriaQuestionReceipt{Schema: "gooo/criteria-questions/v1", Status: "STARTED", SourceSHA256: digest(source),
		ContractSHA256: receipt.ContractSHA256, Options: o, InitialInputs: append([]string(nil), p.spec.Criteria.Inputs...),
		Scope: "source question pool; unlabelled record outputs choose a distinguishing input; only initial and acquired inputs are evaluated by fixed source predicates; all counts are bounded finite observations"}
	for _, criterion := range p.criteria {
		r.Bindings = append(r.Bindings, RecordCriterionBinding{criterion.declaration.ID, criterion.declaration.Activity, criterion.activityID, criterion.goSHA256})
	}
	receipt.CriteriaQuestions = r
	p.effectiveCriteriaInputs = append([]string(nil), p.spec.Criteria.Inputs...)
	defer func() {
		r.EffectiveInputs = len(p.effectiveCriteriaInputs)
		r.EffectiveInputsSHA256 = criteriaInputsSHA(p.effectiveCriteriaInputs)
		if failure != nil {
			r.Status = "FAILED"
		}
	}()
	e, err := newCriteriaQuestionEvaluator(ctx, *p, r)
	if err != nil {
		return err
	}
	for round := 0; ; round++ {
		ranked, err := e.rank(ctx, p.effectiveCriteriaInputs)
		r.Rounds = append(r.Rounds, ranked)
		if err != nil {
			return err
		}
		if ranked.RecommendedIndex == nil {
			r.Status = recordObservationStop(RecordObservationRound{Unobserved: ranked.Unobserved, SurvivingMasks: ranked.SurvivingMasks, UnavailableMasks: ranked.UnavailableMasks})
			if o.ResolveUniqueCandidate && r.Status == "ONE_SURVIVING_CANDIDATE" {
				r.ResolvedMask = new(ranked.SurvivingMasks[0])
			}
			return nil
		}
		if round == o.MaxRounds {
			r.Status = "ROUND_BUDGET"
			return nil
		}
		if len(p.effectiveCriteriaInputs)+len(p.spec.Criteria.HoldoutInputs)+len(p.spec.ValueCases)+len(p.spec.ValueHoldoutCases) >= 128 {
			r.Status = "CASE_BUDGET"
			return nil
		}
		input := ranked.Probes[*ranked.RecommendedIndex].Input
		r.Rounds[len(r.Rounds)-1].AcquiredInput = input
		r.AcquiredInputs = append(r.AcquiredInputs, input)
		p.effectiveCriteriaInputs = append(p.effectiveCriteriaInputs, input)
	}
}

func applyResolvedRecordMask(r *RecordAssemblyReceipt, mask uint16) {
	r.Ranking = []uint16{mask}
	for i := range 1 << len(r.Choices) {
		if uint16(i) != mask {
			r.Ranking = append(r.Ranking, uint16(i))
		}
	}
}
