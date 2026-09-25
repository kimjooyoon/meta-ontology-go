package cache

const (
	// DefaultKeyVersion is the current cache-key schema.
	DefaultKeyVersion    = ProjectionKeyVersion
	ProjectionKeyVersion = "v2"
	keyDomain            = "gooo-projection-key\x00"
)

// ProjectionKeySpec is the typed identity of one reconstructable projection.
// Every field affects the content address; metadata cannot change identity
// after a key has been created.
type ProjectionKeySpec struct {
	Domain                string
	Namespace             string
	Version               string
	ArtifactKind          string
	Projection            string
	HostStage             HostStage
	SemanticClosureDigest Digest
	DependencyRoot        Digest
	PolicySchemaDigest    Digest
	Toolchain             string
	ToolVersion           string
	Target                string
	BuildTags             []string
	OptionsDigest         Digest
	// Options is retained only for source compatibility and is fail-closed.
	Options any
}

// KeySpec is the compatibility constructor for callers that provide source
// inputs and freshness records instead of precomputed typed digests.
type KeySpec struct {
	Version       string
	Domain        string
	Namespace     string
	ArtifactKind  string
	Projection    string
	ToolVersion   string
	Toolchain     string
	Target        string
	BuildTags     []string
	HostStage     HostStage
	Inputs        any
	OptionsDigest Digest
	// Options is retained only for source compatibility and is fail-closed.
	Options   any
	Freshness FreshnessSpec
}

// ProjectionKey is the content address of one projection. Its fields are
// comparable typed identity, so callers can safely use it as a test value.
type ProjectionKey struct {
	Digest                Digest
	Domain                string
	Namespace             string
	Version               string
	ArtifactKind          string
	Projection            string
	HostStage             HostStage
	SemanticClosureDigest Digest
	DependencyRoot        Digest
	PolicySchemaDigest    Digest
	Toolchain             string
	ToolVersion           string
	Target                string
	BuildTagsDigest       Digest
	OptionsDigest         Digest
	InputDigest           Digest
	DependencyDigest      Digest
	ProvenanceDigest      Digest
}
