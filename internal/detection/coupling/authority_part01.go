package coupling

func configFromAuthority(authority AuthorityContext) Config {
	return Config{
		Schema:                  ConfigSchemaV1,
		RegistryDigest:          authority.Registry.Digest,
		ToolchainDigest:         authority.ToolchainDigest,
		ProfileDigest:           authority.ProfileDigest,
		SnapshotDigest:          authority.SnapshotDigest,
		ExpectedProviderDigest:  authority.ExpectedProviderDigest,
		ExpectedObserverDigest:  authority.ExpectedObserverDigest,
		Baseline:                authority.Baseline,
		ExternalReceiptRequired: authority.ExternalReceiptRequired,
	}
}
func normalizeAuthorityContext(authority AuthorityContext) (Config, *evaluationIssue) {
	if authority.Schema == "" {
		return Config{}, required("authority context")
	}
	if authority.Schema != AuthorityContextSchemaV1 {
		return Config{}, failIssue(ReasonMalformedBinding, "authority context schema")
	}
	if !authority.ExternalReceiptRequired {
		return Config{}, unknownIssue(ReasonRequiredInputMissing, "authority resource requirement")
	}
	expected := configFromAuthority(authority)
	if issue := normalizeConfig(expected); issue != nil {
		return Config{}, issue
	}
	if authority.Registry.Surfaces == nil {
		return Config{}, required("authority registry surfaces")
	}
	if _, issue := normalizeRegistry(authority.Registry, expected); issue != nil {
		return Config{}, issue
	}
	if len(authority.Registry.Surfaces) == 0 {
		if issue := validateApplicabilityProof(authority, expected); issue != nil {
			return Config{}, issue
		}
	}
	return expected, nil
}
