package audienceresolution

func buildViews(contract Contract, indicators []Indicator, decision, reason string) []AudienceView {
	views := make([]AudienceView, 0, len(contract.Audiences))
	for _, audience := range contract.Audiences {
		satisfied := 0
		for _, coordinate := range audience.Coordinates {
			for _, indicator := range indicators {
				if indicator.ID == coordinate && indicator.Satisfied {
					satisfied++
				}
			}
		}
		views = append(views, AudienceView{Audience: audience.Audience, Resolution: audience.Resolution,
			Decision: decision, Reason: reason, Satisfied: satisfied, Total: len(audience.Coordinates),
			BasisPoints:            basisPoints(satisfied, len(audience.Coordinates)),
			CoordinateIDs:          append([]string(nil), audience.Coordinates...),
			OmittedCoordinateCount: IndicatorTotal - len(audience.Coordinates)})
	}
	return views
}

func viewCoordinates(contract Contract, audience string) []string {
	for _, value := range contract.Audiences {
		if value.Audience == audience {
			return value.Coordinates
		}
	}
	return nil
}

func coordinateNesting(contract Contract) bool {
	if len(contract.Audiences) != 3 || contract.Audiences[0].Audience != "USER" ||
		contract.Audiences[1].Audience != "TOOL_AUTHOR" || contract.Audiences[2].Audience != "GOVERNOR" {
		return false
	}
	for index := 1; index < len(contract.Audiences); index++ {
		previous, current := contract.Audiences[index-1].Coordinates, contract.Audiences[index].Coordinates
		if len(current) <= len(previous) {
			return false
		}
		for coordinateIndex, coordinate := range previous {
			if current[coordinateIndex] != coordinate {
				return false
			}
		}
	}
	return true
}

func validResolutions(contract Contract) bool {
	return contract.Audiences[0].Resolution == "USER_VISIBLE_COORDINATES" &&
		contract.Audiences[1].Resolution == "TOOL_CONTRACT_COORDINATES" &&
		contract.Audiences[2].Resolution == "GOVERNOR_FULL_LEDGER"
}

func knownAudience(value string) bool {
	return value == "USER" || value == "TOOL_AUTHOR" || value == "GOVERNOR" || value == "all-audiences" || value == "independent.checker"
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
}
