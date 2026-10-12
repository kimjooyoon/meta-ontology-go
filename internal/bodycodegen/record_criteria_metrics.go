package bodycodegen

func recordCriteriaDimensions(r *RecordCriteriaEvidence, prefix string) []CompletenessDimension {
	if r == nil {
		return nil
	}
	evidence := []string{"record_criteria.suite_sha256:" + r.SuiteSHA256, "record_criteria.generated_sha256:" + r.GeneratedSHA256}
	dimensions := []CompletenessDimension{
		completenessDimension(prefix+"_satisfaction", r.Passed, r.Total, "declared predicate/input pairs",
			"Counts true predicates; false and unavailable results remain in the denominator. Scope is the declared finite inputs.", evidence, false),
		completenessDimension(prefix+"_observed", r.Passed+r.Failed, r.Total, "declared predicate/input pairs returning Boolean",
			"Both true and false are observed results. Candidate and predicate evaluation faults remain unknown.", evidence, false),
	}
	for i := range dimensions {
		if r.Unknown == r.Total {
			dimensions[i].Status = "UNKNOWN"
		} else if dimensions[i].Numerator == 0 {
			dimensions[i].Status = "PROGRESS"
		}
	}
	return dimensions
}

func bindRecordCriteriaScope(common *CompletenessReceipt, r *RecordAssemblyReceipt) {
	if r.Criteria == nil {
		return
	}
	c := r.Criteria
	common.Scope["record_criteria"] = map[string]any{
		"passed": c.Passed, "failed": c.Failed, "unknown": c.Unknown, "total": c.Total,
		"selection_order":                "exact-value cases, satisfied predicate/input pairs, matching fields, prior ranking",
		"policy_count_unit":              "exact-value case rows plus criteria input rows satisfying all predicates",
		"held_inputs_used_for_selection": false,
		"model_predicate_body_encoding":  "not present in existing model feature contracts",
	}
}
