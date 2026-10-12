package bodycodegen

func bindCriteriaQuestionScope(common *CompletenessReceipt, r *RecordAssemblyReceipt) {
	o := r.CriteriaQuestions
	if o == nil {
		return
	}
	common.Scope["criteria_questions"] = map[string]any{
		"status": o.Status, "initial_inputs": len(o.InitialInputs), "acquired_inputs": len(o.AcquiredInputs),
		"effective_inputs": o.EffectiveInputs, "effective_inputs_sha256": o.EffectiveInputsSHA256,
		"candidate_evaluations": o.CandidateEvaluations, "predicate_evaluations": o.PredicateEvaluations,
		"unselected_probes_predicate_scored": false,
	}
}
