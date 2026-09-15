package selectiveci

func normalizeSnapshot(s Snapshot) (Snapshot, error) {
	if s.Schema != SchemaV1 {
		return Snapshot{}, fail(CodeInvalidSchema, "schema %q is not %q", s.Schema, SchemaV1)
	}
	if s.Status != StatusBound || s.FullSuiteFallback || s.Digest == "" {
		return Snapshot{}, fail(CodeInvalidStatus, "snapshot must be BOUND with a digest")
	}
	if s.RegisteredIDs == nil || s.Sources == nil {
		return Snapshot{}, fail(CodeInvalidSchema, "registered IDs and sources must be explicit arrays")
	}
	registered, err := normalizeRegisteredIDs(s.RegisteredIDs)
	if err != nil {
		return Snapshot{}, err
	}
	sourceMapDigest, err := normalizeDigest(s.SourceMapDigest, "source-map digest")
	if err != nil {
		return Snapshot{}, err
	}
	registryDigest, err := normalizeDigest(s.RegistryDigest, "registry digest")
	if err != nil {
		return Snapshot{}, err
	}
	sources, err := normalizeManifestSources(s.Sources, registered)
	if err != nil {
		return Snapshot{}, err
	}
	if !validDigest(s.Digest) {
		return Snapshot{}, fail(CodeMalformedDigest, "snapshot digest is malformed")
	}
	s.RegisteredIDs, s.SourceMapDigest = sortedIDs(registered), sourceMapDigest
	s.RegistryDigest, s.Sources = registryDigest, sources
	unsigned, err := s.unsignedJSON()
	if err != nil {
		return Snapshot{}, err
	}
	if digest(unsigned) != s.Digest {
		return Snapshot{}, fail(CodeStaleSnapshot, "snapshot digest does not match canonical content")
	}
	return s, nil
}
