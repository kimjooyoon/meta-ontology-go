package languageexampleexperiment

func closedReport(input Input, reason, resolution string) Report {
	total := input.Contract.Fixed.Indicators
	if total < 1 {
		total = 15
	}
	return finishReport(Report{
		Schema: ReportSchema, Decision: "FAIL_CLOSED", Resolution: resolution, Reason: reason,
		Interpretation: "NO_LANGUAGE_QUALITY_CLAIM", SubjectSHA: input.Profile.SubjectSHA,
		ContractID: input.Contract.ID,
		Summary: Summary{Coordinates: Coordinates{Total: total}, Unknowns: 1,
			NotClaimed: len(input.Contract.NotClaimed)},
		Indicators: []Indicator{}, Views: []View{}, Proofs: []Proof{},
		NotClaimed: append([]string{}, input.Contract.NotClaimed...),
	})
}
