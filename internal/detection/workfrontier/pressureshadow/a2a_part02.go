package pressureshadow

func selectorPathIDs(input Input) map[string]struct{} {
	ids := make(map[string]struct{}, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		ids[pathID(path)] = struct{}{}
	}
	return ids
}
func coverageRows(input Input) map[string]PathCoverage {
	rows := make(map[string]PathCoverage, len(input.PathCoverage))
	for _, row := range input.PathCoverage {
		rows[row.PathID] = row
	}
	return rows
}
func missingPathIDs(paths map[string]struct{}, rows map[string]PathCoverage) []string {
	missing := []string{}
	for id := range paths {
		if _, exists := rows[id]; !exists {
			missing = append(missing, id)
		}
	}
	return missing
}
func orphanPathIDs(paths map[string]struct{}, rows map[string]PathCoverage) []string {
	orphan := []string{}
	for id := range rows {
		if _, exists := paths[id]; !exists {
			orphan = append(orphan, id)
		}
	}
	return orphan
}
func bindingIssues(input Input, paths map[string]struct{}, rows map[string]PathCoverage) ([]string, []string) {
	missing, mismatch := []string{}, []string{}
	selector := []string{input.Selector.SnapshotDigest, input.Selector.PolicyDigest, input.Selector.RegistryDigest}
	for id := range paths {
		row, exists := rows[id]
		if !exists {
			continue
		}
		values := []string{row.SnapshotDigest, row.PolicyDigest, row.RegistryDigest}
		blank, unequal := false, false
		for index := range selector {
			blank = blank || selector[index] == "" || values[index] == ""
			unequal = unequal || selector[index] != "" && values[index] != "" &&
				values[index] != selector[index]
		}
		if blank {
			missing = append(missing, id)
		}
		if unequal {
			mismatch = append(mismatch, id)
		}
	}
	return missing, mismatch
}
func pathDecision(pathCount int, missing, orphan, missingBinding, mismatch []string) (Decision, Reason) {
	switch {
	case len(orphan) > 0:
		return DecisionFailClosed, ReasonOrphanPathCoverage
	case pathCount == 0:
		return DecisionUnknown, ReasonRequiredInputMissing
	case len(missing) > 0:
		return DecisionUnknown, ReasonMissingPathCoverage
	case len(missingBinding) > 0:
		return DecisionUnknown, ReasonRequiredInputMissing
	case len(mismatch) > 0:
		return DecisionUnknown, ReasonBindingMismatch
	default:
		return DecisionPass, ReasonNone
	}
}
