package languagepackageruntime

import "github.com/kimjooyoon/meta-ontology-go/internal/meta/languageconcept"

type Input struct {
	ExpectedHeadSHA string
	ConceptArtifact languageconcept.Artifact
	RegistryRaw     []byte
}
