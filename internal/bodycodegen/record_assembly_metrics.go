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
	for i := range dimensions {
		if dimensions[i].Numerator == 0 && dimensions[i].Denominator > 0 {
			dimensions[i].Status = "PROGRESS"
		}
	}
	return dimensions
}
