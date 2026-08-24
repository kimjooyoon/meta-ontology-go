package languagedelivery

func groupByClass(results []ObligationResult) []NamedCoordinates {
	classes := []IndicatorClass{ClassOutcome, ClassDriver, ClassGuardrail}
	groups := make([]NamedCoordinates, 0, len(classes))
	for _, class := range classes {
		var selected []ObligationResult
		for _, result := range results {
			if result.Class == class {
				selected = append(selected, result)
			}
		}
		groups = append(groups, NamedCoordinates{Name: string(class), Coordinates: coordinates(selected)})
	}
	return groups
}

func groupByOwner(results []ObligationResult) []NamedCoordinates {
	groups := make([]NamedCoordinates, 0, len(audienceOrder))
	for _, audience := range audienceOrder {
		var selected []ObligationResult
		for _, result := range results {
			if result.Audience == audience {
				selected = append(selected, result)
			}
		}
		groups = append(groups, NamedCoordinates{Name: string(audience), Coordinates: coordinates(selected)})
	}
	return groups
}

func internalReadiness(receipt ReadinessArtifact) InternalReadiness {
	value := InternalReadiness{Claim: "INTERNAL_SELF_IMPROVEMENT_CONTRACT", Total: len(receipt.Report.Obligations)}
	for _, item := range receipt.Report.Obligations {
		if item.Status == "SATISFIED" {
			value.Satisfied++
		}
	}
	if value.Total != 0 {
		value.BasisPoints = value.Satisfied * 10000 / value.Total
	}
	return value
}
