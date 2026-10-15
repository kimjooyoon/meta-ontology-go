package bodycodegen

func recordAssemblyDimensions(r *RecordAssemblyReceipt) []CompletenessDimension {
	dimensions := []CompletenessDimension{
		completenessDimension("declared_record_case_accuracy", r.Passed, r.Total, "declared typed selection cases",
			"Matches complete named record results in the bounded pure-value evaluator; native runtime cases are observed separately.",
			[]string{"record_assembly.test_suite_sha256:" + r.TestSuiteSHA256}, false),
		completenessDimension("declared_record_field_accuracy", r.FieldsPassed, r.FieldsTotal, "declared expected record output fields",
			"Counts each expected field in the supplied selection cases; preserves partially matching record outputs.",
			[]string{"record_assembly.contract_sha256:" + r.ContractSHA256}, false),
	}
	if r.Schema == scalarCriteriaAssemblySchema {
		dimensions = dimensions[:1]
		dimensions[0].ID = "declared_scalar_case_accuracy"
		dimensions[0].Unit = "declared Integer selection cases"
		dimensions[0].Reason = "Matches supplied scalar expectations; predicate/input obligations have separate counts."
	}
	if o := r.Observation; o != nil {
		dimensions[0].Unit = "original and acquired typed selection cases"
		dimensions[0].Reason = "Counts original cases plus retained oracle answers; held and native cases remain separate."
		dimensions[0].Evidence = append(dimensions[0].Evidence, "record_observation.effective_cases_sha256:"+o.EffectiveCasesSHA256)
		dimensions[1].Unit = "fields in original and acquired selection cases"
		dimensions[1].Reason = "Counts fields from the effective selection suite, including retained oracle answers."
	}
	for i := range dimensions {
		if dimensions[i].Numerator == 0 && dimensions[i].Denominator > 0 {
			dimensions[i].Status = "PROGRESS"
		}
	}
	if r.Criteria != nil && r.Total == 0 {
		dimensions = nil
	}
	if r.Holdout != nil {
		dimensions = append(dimensions, recordHoldoutDimensions(r.Holdout)...)
	}
	dimensions = append(dimensions, recordCriteriaDimensions(r.Criteria, "criteria_selection")...)
	dimensions = append(dimensions, recordCriteriaDimensions(r.CriteriaHoldout, "criteria_holdout")...)
	return dimensions
}
