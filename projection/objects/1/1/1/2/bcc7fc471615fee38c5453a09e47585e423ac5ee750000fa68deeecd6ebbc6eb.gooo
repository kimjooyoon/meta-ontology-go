package bidir

import (
	"errors"
	"fmt"
	"strings"
)

// MeasureBXFixture runs the hard evidence contract for one fixture.
func MeasureBXFixture(fixture ReconciliationFixture) (BXEvidence, error) {
	if fixture == nil {
		return BXEvidence{}, errors.New("reconciliation fixture is nil")
	}
	contract, ok := fixture.(BXEvidenceFixture)
	if !ok {
		return BXEvidence{}, errors.New("fixture does not implement the hard BX evidence contract")
	}
	evidence := BXEvidence{SchemaVersion: BXEvidenceSchemaVersion, Fixture: strings.TrimSpace(fixture.Name())}
	if evidence.Fixture == "" {
		return BXEvidence{}, errors.New("reconciliation fixture name is empty")
	}
	document := fixture.Document()
	base, err := Get(document)
	if err != nil {
		return evidence, fmt.Errorf("get base document: %w", err)
	}
	evidence.Base, err = baseEvidence(contract.BaseEvidence(), document, base)
	if err != nil {
		return evidence, err
	}
	evidence.GetPutPassed = CheckGetPut(document) == nil
	acceptedDelta := fixture.AcceptedDelta()
	accepted, err := Reconcile(base, acceptedDelta)
	if err != nil {
		return evidence, fmt.Errorf("reconcile accepted delta: %w", err)
	}
	evidence.AcceptedRelationAdds = len(accepted.Delta.AddedRelations)
	evidence.Locality = accepted.Locality
	evidence.Delta, err = makeDeltaEvidence(acceptedDelta, accepted.Locality, false, base, accepted.Model)
	if err != nil {
		return evidence, fmt.Errorf("accepted delta evidence: %w", err)
	}
	updatedDocument, err := Put(document, accepted.Model)
	if err != nil {
		return evidence, fmt.Errorf("put accepted model: %w", err)
	}
	observed, err := Get(updatedDocument)
	if err != nil {
		return evidence, fmt.Errorf("get updated document: %w", err)
	}
	evidence.PutGetPassed = true
	evidence.SemanticEquivalent = SemanticEquivalent(accepted.Model, observed)
	evidence.AcceptedTransaction, err = acceptedTransaction(contract, document, updatedDocument, base, accepted)
	if err != nil {
		return evidence, err
	}
	observer, err := contract.RejectedWriteObserver(document)
	if err != nil {
		return evidence, fmt.Errorf("rejected write observer: %w", err)
	}
	evidence.PartialConflict, evidence.RejectedTransaction, evidence.PartialDelta, err = partialEvidence(document, base, fixture.PartialDelta(), observer)
	if err != nil {
		return evidence, err
	}
	evidence.Deferred = deferredBXSeams()
	if err := evidence.validate(); err != nil {
		return evidence, err
	}
	return evidence, nil
}
