package workfrontier

func r4RequiredInputKnown(input R4Input) bool {
	return input.SchemaVersion == R4SchemaVersion && input.SnapshotDigest != "" &&
		input.SnapshotPayload != "" && input.PolicyDigest != "" &&
		input.PolicyPayload != "" && input.RegistryDigest != "" &&
		input.RegistryPayload != "" &&
		input.MinimumSelectedPressures >= 2 && input.Pressures != nil &&
		input.States != nil && input.Paths != nil && input.RootObligationIDs != nil &&
		len(input.RootObligationIDs) != 0 && input.Rules != nil
}
func r4StableDeclarationsKnown(input R4Input) bool {
	for _, pressure := range input.Pressures {
		if pressure.StableID == "" {
			return false
		}
	}
	for _, state := range input.States {
		if state.ObligationID == "" || state.Status == "" {
			return false
		}
	}
	for _, path := range input.Paths {
		if path.StableID == "" || path.ObligationID == "" {
			return false
		}
	}
	return true
}
