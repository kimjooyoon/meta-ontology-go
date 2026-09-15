package artifactresolutionexperiment

func closed(input Input, resolution, reason string) Report {
	facts := digestValue(input)
	return seal(Report{
		Schema: ReportSchema, Decision: "FAIL_CLOSED", Resolution: resolution,
		Reason: reason, Interpretation: "NO_LANGUAGE_QUALITY_CLAIM",
		SubjectSHA: input.SubjectSHA, ContractID: input.Contract.ID,
		Summary: Summary{Coordinates: Coordinates{Satisfied: 0, Total: ExpectedIndicators},
			NotClaimed: len(input.Contract.NotClaimed), Unknowns: 1},
		Indicators: []Indicator{}, Views: []View{}, Proofs: []Proof{},
		NotClaimed: input.Contract.NotClaimed, FactsDigest: facts,
	})
}
