package ciplanusecase

func Evaluate(input Input) Report {
	summary := summarize(input)
	cases := make([]CaseResult, 0, len(input.Contract.Cases))
	for _, spec := range input.Contract.Cases {
		result, satisfied := evaluateCase(spec, input)
		cases = append(cases, result)
		if satisfied {
			summary.CasesSatisfied++
		}
	}
	indicators := buildIndicators(summary, input.Contract.Limits)
	proofs := buildProofs(summary, input.Contract)
	decision := "PASS"
	for _, indicator := range indicators {
		if indicator.Status != "SATISFIED" {
			decision = "FAIL_CLOSED"
		}
	}
	for _, proof := range proofs {
		if proof.Status != "SATISFIED" {
			decision = "FAIL_CLOSED"
		}
	}
	report := Report{
		Schema: ReportSchema, Decision: decision, Resolution: "EXACT",
		Interpretation: "MINIMAL_CI_PLAN_VALUE_OBSERVED", ContractDigest: digest(input.Contract),
		Cases: cases, Summary: summary, Indicators: indicators, ReaderViews: buildReaderViews(indicators),
		Proofs: proofs, NotClaimed: append([]string(nil), input.Contract.NotClaimed...),
	}
	return seal(report)
}
