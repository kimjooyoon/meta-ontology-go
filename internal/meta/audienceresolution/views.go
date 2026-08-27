package audienceresolution

func buildViews(model semanticSourceModel, state recordState, globalDecision, globalResolution, globalReason string) []AudienceView {
	views := make([]AudienceView, 0, len(model.Audiences))
	required := len(sourceCoordinates(model))
	for _, audience := range model.Audiences {
		satisfied := 0
		visible := 0
		omitted := make([]OmittedEvidence, 0)
		contradictory := ""
		for _, coordinate := range audience.Coordinates {
			if coordinateVisible(state, coordinate) {
				visible++
			}
			if state.valid[coordinate] {
				satisfied++
			}
			if coordinateContradictory(state, coordinate) && contradictory == "" {
				contradictory = coordinate
			}
		}
		for _, coordinate := range sourceCoordinates(model) {
			if !contains(audience.Coordinates, coordinate) || !coordinateVisible(state, coordinate) {
				omitted = append(omitted, omittedEvidence(coordinate))
			}
		}
		localDecision, localResolution, localReason := localState(audience, state, required, omitted, contradictory)
		inherited := "LOCALLY_VERIFIED"
		if localDecision != globalDecision {
			inherited = "INHERITED_NOT_LOCALLY_VERIFIED"
		}
		views = append(views, AudienceView{Audience: audience.Audience, Resolution: audience.Resolution,
			GlobalDecision: globalDecision, InheritedStatus: inherited, LocalDecision: localDecision,
			LocalResolution: localResolution, LocalReason: localReason, OmittedEvidence: omitted,
			Satisfied: satisfied, Total: len(audience.Coordinates), Visible: visible, Required: required,
			BasisPoints: basisPoints(satisfied, len(audience.Coordinates)), CoordinateIDs: append([]string(nil), audience.Coordinates...),
			OmittedCoordinateCount: required - visible})
	}
	return views
}

func localState(audience AudienceContract, state recordState, required int, omitted []OmittedEvidence, contradictory string) (string, string, string) {
	if contradictory != "" {
		return "REFUTED", "INVARIANT_ONLY", "VISIBLE_EVIDENCE_CONTRADICTION:" + omittedOrCoordinate(state, contradictory)
	}
	for _, coordinate := range audience.Coordinates {
		if !coordinateVisible(state, coordinate) || !state.valid[coordinate] {
			item := omittedEvidence(coordinate)
			return "UNKNOWN", "LOWER_RESOLUTION", "VISIBLE_EVIDENCE_INSUFFICIENT:" + item.Stage + ":" + item.Step + ":" + item.Reason
		}
	}
	if len(omitted) > 0 || len(audience.Coordinates) < required {
		item := omitted[0]
		return "UNKNOWN", "LOWER_RESOLUTION", "REQUIRED_EVIDENCE_OMITTED:" + item.Stage + ":" + item.Step + ":" + item.Reason
	}
	return "PASS", "EXACT", "LOCAL_EVIDENCE_COMPLETE"
}

func omittedOrCoordinate(state recordState, coordinate string) string {
	if record, ok := state.records[coordinate]; ok {
		return record.Stage + ":" + record.Step + ":" + record.Reason
	}
	return coordinate
}

func omittedEvidence(coordinate string) OmittedEvidence {
	for _, spec := range indicatorSpecs() {
		if spec.ID == coordinate {
			return OmittedEvidence{Coordinate: coordinate, Stage: spec.Stage, Step: spec.Step, Reason: spec.Reason}
		}
	}
	return OmittedEvidence{Coordinate: coordinate, Stage: "projection", Step: "policy", Reason: "formal source policy coordinate has no visible raw observation"}
}

func basisPoints(satisfied, total int) int {
	if total == 0 {
		return 0
	}
	return satisfied * 10000 / total
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
	if len(contract.Audiences) != 3 {
		return false
	}
	previous := []string{}
	for _, audience := range contract.Audiences {
		if len(audience.Coordinates) <= len(previous) {
			return false
		}
		for index, coordinate := range previous {
			if audience.Coordinates[index] != coordinate {
				return false
			}
		}
		previous = audience.Coordinates
	}
	return true
}

func validResolutions(contract Contract) bool {
	return len(contract.Audiences) == 3 && contract.Audiences[0].Resolution == "USER_VISIBLE_COORDINATES" &&
		contract.Audiences[1].Resolution == "TOOL_CONTRACT_COORDINATES" && contract.Audiences[2].Resolution == "GOVERNOR_FULL_LEDGER"
}

func knownAudience(value string) bool {
	return value == "USER" || value == "TOOL_AUTHOR" || value == "GOVERNOR" || value == "all-audiences" || value == "independent.checker"
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
