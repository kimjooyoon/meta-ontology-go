package pressurecoverage

func fixture() Input {
	input := Input{
		Schema:                  SchemaVersion,
		RequestedK:              21,
		MinimumIndependent:      2,
		AuthoritySnapshotDigest: expectedSnapshot,
		PolicyDigest:            expectedPolicy,
		RegistryDigest:          expectedRegistry,
		ToolchainOptionsDigest:  expectedToolchain,
		PressureRecords: []PressureRecord{
			{"pressure-z", "category-z", "group-z", "rule-1"},
			{"pressure-a", "category-a", "group-a", "rule-1"},
			{"pressure-b", "category-b", "group-b", "rule-1"},
			{"pressure-aa", "category-a", "group-a", "rule-1"},
		},
		RequiredPressureIDs: []string{"pressure-z", "pressure-a", "pressure-b", "pressure-aa"},
	}
	return input
}
func bindingField(input Input, role string) string {
	switch role {
	case "authority-snapshot":
		return input.AuthoritySnapshotDigest
	case "policy":
		return input.PolicyDigest
	case "registry":
		return input.RegistryDigest
	case "toolchain-options":
		return input.ToolchainOptionsDigest
	default:
		return ""
	}
}
func schemaJSON(suffix string) string {
	return `{"schema":"` + SchemaVersion + `"` + suffix + `}`
}
