package toolchainformatfix

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"
	cliruntime "github.com/kimjooyoon/meta-ontology-go/internal/toolchaincli"
)

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact languageconcept.Artifact
	RegistryRaw     []byte
	Executor        cliruntime.Executor
}
