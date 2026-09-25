package coupling

func validateManifestHeader(manifest ChangeManifest, config Config) *evaluationIssue {
	if manifest.Schema == "" {
		return required("manifest")
	}
	if manifest.Schema != ManifestSchemaV1 {
		return failIssue(ReasonMalformedBinding, "manifest schema")
	}
	if !manifest.Complete {
		return required("complete source-backed change manifest")
	}
	for _, value := range []struct {
		value string
		name  string
	}{
		{manifest.RegistryDigest, "manifest registry digest"},
		{manifest.ToolchainDigest, "manifest toolchain digest"},
		{manifest.ProfileDigest, "manifest profile digest"},
		{manifest.BeforeSnapshotDigest, "manifest before snapshot digest"},
		{manifest.AfterSnapshotDigest, "manifest after snapshot digest"},
		{manifest.Digest, "manifest digest"},
	} {
		if issue := normalizeDigestValue(value.value, value.name); issue != nil {
			return issue
		}
	}
	if manifest.RegistryDigest != config.RegistryDigest || manifest.ToolchainDigest != config.ToolchainDigest ||
		manifest.ProfileDigest != config.ProfileDigest || manifest.AfterSnapshotDigest != config.SnapshotDigest {
		return failIssue(ReasonDigestMismatch, "manifest/config digest")
	}
	if stableDigest(manifestCanonical(manifest)) != manifest.Digest {
		return unknownIssue(ReasonStaleInput, "manifest digest")
	}
	if manifest.Entries == nil {
		return required("complete manifest entries")
	}
	return nil
}
