package languagereadiness

import (
	"fmt"

	conceptoperation "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

// EvaluateWithConceptOperation adds only independently verified concept-operation
// evidence to the ordinary readiness evaluator. It does not authorize promotion.
func EvaluateWithConceptOperation(raw []byte, evidence conceptoperation.Receipt) (Snapshot, error) {
	if err := conceptoperation.VerifyReceipt(evidence, "", ""); err != nil {
		return Snapshot{}, fmt.Errorf("verify concept-operation evidence: %w", err)
	}
	if evidence.Status != "VERIFIED" {
		return Snapshot{}, fmt.Errorf("concept-operation evidence is not verified")
	}
	return evaluate(raw, evidenceDigests{conceptOperation: evidence.Digest})
}
