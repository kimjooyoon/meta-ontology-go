package languagereadiness

import (
	"fmt"
	"os"

	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

// EvaluateWithConceptOperation adds only independently verified concept-operation
// evidence to the ordinary readiness evaluator. It does not authorize promotion.
func EvaluateWithConceptOperation(raw, evidenceRaw []byte, inputs conceptoperation.SourceInputs, repositoryRoot, expectedRepository, expectedSubjectSHA string) (Snapshot, error) {
	if repositoryRoot == "" || expectedRepository == "" || expectedSubjectSHA == "" || inputs.ScratchDirectory == "" {
		return Snapshot{}, fmt.Errorf("concept-operation observation requires exact repository, head, root, and external scratch")
	}
	evidence, err := conceptoperation.VerifyEvidence(evidenceRaw, inputs, os.DirFS(repositoryRoot),
		repositoryRoot, expectedRepository, expectedSubjectSHA)
	if err != nil {
		return Snapshot{}, err
	}
	if evidence.Status != "VERIFIED" {
		return Snapshot{}, fmt.Errorf("concept-operation evidence is not verified")
	}
	return evaluate(raw, evidenceDigests{conceptOperation: evidence.Digest})
}
