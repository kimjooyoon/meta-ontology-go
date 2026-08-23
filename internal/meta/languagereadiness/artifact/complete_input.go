package artifact

type CompleteEvidenceInput struct {
	ConceptArtifact []byte
	Promotion       []byte
	Capability      []byte
	UseCases        []byte
	Syntax          []byte
	Diagnostic      []byte
	PackageRuntime  []byte
	ToolchainCLI    []byte
	HeadSHA         string
}
