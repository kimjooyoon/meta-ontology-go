package artifact

import (
	"fmt"
	"os"

	readiness "github.com/kimjooyoon/meta-ontology-go/internal/meta/languagereadiness"
	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

// BuildWithConceptOperationEvidence observes one verified concept-operation
// binding without entering the promotion or complete-evidence paths.
func BuildWithConceptOperationEvidence(
	conceptArtifact []byte,
	conceptOperationRaw []byte,
	conceptOperationInputs conceptoperation.SourceInputs,
	repositoryRoot, scratchDirectory, expectedRepository, headSHA string,
) (Receipt, error) {
	if repositoryRoot == "" || scratchDirectory == "" || expectedRepository == "" || headSHA == "" {
		return Receipt{}, fmt.Errorf("concept-operation observation requires exact repository, head, and external scratch")
	}
	conceptOperation, err := conceptoperation.Verify(conceptOperationRaw, expectedRepository, headSHA)
	if err != nil {
		return Receipt{}, err
	}
	conceptOperationInputs.ScratchDirectory = scratchDirectory
	if err := conceptoperation.VerifySource(
		conceptOperation, conceptOperationInputs, os.DirFS(repositoryRoot), repositoryRoot,
		expectedRepository, headSHA,
	); err != nil {
		return Receipt{}, err
	}
	snapshot, err := readiness.EvaluateWithConceptOperation(conceptArtifact, conceptOperation)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, "", "")
}
