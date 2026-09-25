package userjourneyscorecard

func expectedContract() Contract {
	contract := expectedContractBase()
	for _, journey := range contract.Journeys {
		if journey.ID == "run-package" {
			panic("userjourneyscorecard: run-package denominator duplicated")
		}
	}
	contract.Journeys = append(contract.Journeys, JourneyDefinition{
		ID:            "run-package",
		Operation:     "execute-multi-file-package",
		Arguments:     []string{"run", "--entry", "PayOrder", "examples/billing-package"},
		ProofChoice:   "FOUNDATION",
		MetaOperation: "bind-package-files-then-execute-activity",
	})
	return contract
}
