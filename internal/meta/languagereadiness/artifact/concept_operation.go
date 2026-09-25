package artifact

import (
	"fmt"

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
	conceptOperationInputs.ScratchDirectory = scratchDirectory
	snapshot, err := readiness.EvaluateWithConceptOperation(
		conceptArtifact, conceptOperationRaw, conceptOperationInputs, repositoryRoot,
		expectedRepository, headSHA,
	)
	if err != nil {
		return Receipt{}, err
	}
	return build(snapshot, headSHA, "", "")
}
