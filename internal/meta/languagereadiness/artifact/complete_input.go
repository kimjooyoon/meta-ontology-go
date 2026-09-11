package artifact

import conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"

type CompleteEvidenceInput struct {
	ConceptArtifact        []byte
	ConceptOperationBinding []byte
	ConceptOperationInputs conceptoperation.SourceInputs
	RepositoryRoot         string
	ConceptOperationScratchDirectory string
	Promotion              []byte
	Capability             []byte
	UseCases               []byte
	Syntax                 []byte
	Diagnostic             []byte
	PackageRuntime         []byte
	ToolchainCLI           []byte
	ToolchainFormatFix     []byte
	ToolchainLSP           []byte
	ToolchainConformance   []byte
	ToolchainRelease       []byte
	ExpectedRepository     string
	HeadSHA                string
	ExpectedPredecessorSHA string
}
