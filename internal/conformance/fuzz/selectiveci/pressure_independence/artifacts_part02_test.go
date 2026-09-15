package pressureindependence

func bindSyntheticArtifacts(input *Input) {
	input.AuthoritySnapshotDigest = artifactDigest(authoritySnapshotArtifactBytes)
	input.PolicyDigest = artifactDigest(policyArtifactBytes)
	input.RegistryDigest = artifactDigest(pressureRegistryArtifactBytes)
	input.OracleDigest = artifactDigest(oracleContractArtifactBytes)
	input.ToolchainOptionsDigest = artifactDigest(toolchainOptionsArtifactBytes)
}
func syntheticArtifactRoles() []artifactRole {
	return []artifactRole{
		{
			name: "authority_snapshot", data: authoritySnapshotArtifactBytes,
			set: func(input *Input, digest string) { input.AuthoritySnapshotDigest = digest },
			get: func(input Input) string { return input.AuthoritySnapshotDigest },
		},
		{
			name: "policy", data: policyArtifactBytes,
			set: func(input *Input, digest string) { input.PolicyDigest = digest },
			get: func(input Input) string { return input.PolicyDigest },
		},
		{
			name: "pressure_registry", data: pressureRegistryArtifactBytes,
			set: func(input *Input, digest string) { input.RegistryDigest = digest },
			get: func(input Input) string { return input.RegistryDigest },
		},
		{
			name: "oracle_contract", data: oracleContractArtifactBytes,
			set: func(input *Input, digest string) { input.OracleDigest = digest },
			get: func(input Input) string { return input.OracleDigest },
		},
		{
			name: "toolchain_options", data: toolchainOptionsArtifactBytes,
			set: func(input *Input, digest string) { input.ToolchainOptionsDigest = digest },
			get: func(input Input) string { return input.ToolchainOptionsDigest },
		},
	}
}
