package selectiveci

import (
	"fmt"
)

func fail(code ErrorCode, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), FullSuiteFallback: true}
}

// NewSnapshot constructs a canonical source-backed snapshot from explicit
// semanticbinding records.
func NewSnapshot(input SnapshotInput) (Snapshot, error) { return Build(input) }

// Build is the primary Snapshot constructor.
func Build(input SnapshotInput) (Snapshot, error) {
	if input.RegisteredIDs == nil {
		return unknownSnapshot(fail(CodeInput, "registered IDs are required"))
	}
	if len(input.CandidateBindings) != 0 {
		return unknownSnapshot(fail(CodeCandidateIdentity, "candidate-only bindings cannot become authoritative"))
	}
	if len(input.DerivedBindings) != 0 {
		return unknownSnapshot(fail(CodeDerivedIdentity, "derived-only bindings cannot become authoritative"))
	}
	registered, err := normalizeRegisteredIDs(input.RegisteredIDs)
	if err != nil {
		return unknownSnapshot(err)
	}
	sourceMapDigest, err := normalizeDigest(input.SourceMapDigest, "source-map digest")
	if err != nil {
		return unknownSnapshot(err)
	}
	registryDigest, err := normalizeDigest(input.RegistryDigest, "registry digest")
	if err != nil {
		return unknownSnapshot(err)
	}
	sources, err := normalizeSources(input.Sources, registered)
	if err != nil {
		return unknownSnapshot(err)
	}
	result := Snapshot{
		Schema: SchemaV1, Status: StatusBound,
		SourceMapDigest: sourceMapDigest, RegistryDigest: registryDigest,
		RegisteredIDs: sortedIDs(registered), Sources: sources,
	}
	unsigned, err := result.unsignedJSON()
	if err != nil {
		return unknownSnapshot(err)
	}
	result.Digest = digest(unsigned)
	return result, nil
}

// BuildSnapshot is the descriptive spelling of Build.
func BuildSnapshot(input SnapshotInput) (Snapshot, error) { return Build(input) }
