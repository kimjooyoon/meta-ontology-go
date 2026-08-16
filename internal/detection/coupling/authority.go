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

func validateApplicabilityProof(authority AuthorityContext, expected Config) *evaluationIssue {
	proof := authority.Applicability
	if proof == nil {
		return required("empty-registry applicability proof")
	}
	if proof.Schema == "" {
		return required("applicability proof")
	}
	if proof.Schema != AuthorityContextSchemaV1 {
		return failIssue(ReasonMalformedBinding, "applicability proof schema")
	}
	if !proof.AllowsEmpty {
		return required("empty-registry applicability proof")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{proof.RegistryDigest, "applicability registry digest"},
		{proof.ToolchainDigest, "applicability toolchain digest"},
		{proof.ProfileDigest, "applicability profile digest"},
		{proof.SnapshotDigest, "applicability snapshot digest"},
		{proof.Digest, "applicability proof digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if proof.RegistryDigest != authority.Registry.Digest || proof.ToolchainDigest != expected.ToolchainDigest ||
		proof.ProfileDigest != expected.ProfileDigest || proof.SnapshotDigest != expected.SnapshotDigest {
		return unknownIssue(ReasonStaleInput, "applicability proof")
	}
	if stableDigest(applicabilityCanonical(*proof)) != proof.Digest {
		return unknownIssue(ReasonStaleInput, "applicability proof digest")
	}
	return nil
}

func comparePacketAuthority(input Input, expected Config, authority AuthorityContext) *evaluationIssue {
	if configCanonical(input.Config) != configCanonical(expected) {
		return failIssue(ReasonAuthorityInputSelfBound, "producer config differs from evaluator authority")
	}
	if registryCanonical(input.Registry) != registryCanonical(authority.Registry) {
		return failIssue(ReasonAuthorityInputSelfBound, "producer registry differs from evaluator authority")
	}
	return nil
}

func validatePacketApplicability(input Input, authority AuthorityContext) *evaluationIssue {
	packetRegistryEmpty := input.Registry.Surfaces != nil && len(input.Registry.Surfaces) == 0
	packetManifestEmpty := input.Manifest.Entries != nil && len(input.Manifest.Entries) == 0 && input.Manifest.ZeroChange
	if !packetRegistryEmpty && !packetManifestEmpty {
		return nil
	}
	if len(authority.Registry.Surfaces) != 0 {
		return unknownIssue(ReasonRequiredInputMissing, "empty packet applicability proof")
	}
	return nil
}
