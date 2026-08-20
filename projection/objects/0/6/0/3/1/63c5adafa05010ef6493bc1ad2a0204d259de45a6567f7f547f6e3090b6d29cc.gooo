package pressureindependence

func validate(input Input) validation {
	if missingInput(input) {
		if len(input.FinitePathIDs) == 0 && len(input.GuardIDs) > 0 {
			return validation{decision: DecisionUnknown, reason: ReasonProvPathMissing}
		}
		return validation{decision: DecisionUnknown, reason: ReasonRequiredInputMissing}
	}
	if staleDigest(input) {
		return validation{decision: DecisionUnknown, reason: ReasonStaleDigest}
	}
	if hasDuplicate(input.RequiredPressureIDs) {
		return validation{decision: DecisionFailClosed, reason: ReasonDuplicatePressureID}
	}
	if malformedLists(input) {
		return validation{decision: DecisionFailClosed, reason: ReasonProvPathMalformed}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if _, exists := records[record.PressureID]; exists {
			if records[record.PressureID] == record {
				return validation{decision: DecisionFailClosed, reason: ReasonDuplicatePressureID}
			}
			return validation{decision: DecisionFailClosed, reason: ReasonConflictingGroupBinding}
		}
		records[record.PressureID] = record
	}
	applicabilityRule := ""
	for _, id := range input.RequiredPressureIDs {
		if _, exists := records[id]; !exists {
			return validation{decision: DecisionFailClosed, reason: ReasonCatalogMismatch}
		}
		record := records[id]
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return validation{decision: DecisionUnknown, reason: ReasonApplicabilityUnproven}
		}
		if applicabilityRule == "" {
			applicabilityRule = record.ApplicabilityRuleID
		} else if applicabilityRule != record.ApplicabilityRuleID {
			return validation{decision: DecisionUnknown, reason: ReasonInputAmbiguous}
		}
	}
	return validation{records: records}
}
func missingInput(input Input) bool {
	values := []string{input.Schema, input.FixtureID, input.AuthoritySnapshotDigest, input.PolicyDigest,
		input.RegistryDigest, input.OracleDigest, input.ToolchainOptionsDigest}
	for _, value := range values {
		if !validToken(value) {
			return true
		}
	}
	return input.Schema != SchemaV1 || input.RequestedK == 0 || input.MinimumIndependent == 0 ||
		len(input.PressureRecords) == 0 || len(input.RequiredPressureIDs) == 0 ||
		len(input.GuardIDs) == 0 || len(input.FinitePathIDs) == 0 || missingCeiling(input.ResourceCeilings)
}
func staleDigest(input Input) bool {
	return !verifyArtifactBindings(input)
}
