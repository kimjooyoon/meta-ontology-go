package bidir

import (
	"errors"
	"fmt"
)

// AdaptBXTypedEnvelope joins typed laws and the existing BX evidence bundle.
func AdaptBXTypedEnvelope(projection BXTypedProjectionEnvelope, observation BXTypedObservationEnvelope) (BXTypedEnvelopeResult, error) {
	result := BXTypedEnvelopeResult{SchemaVersion: BXTypedEnvelopeSchemaVersion, FeatureGreen: false, Deferred: deferredBXSeams()}
	if err := projection.validate(); err != nil {
		return result, fmt.Errorf("typed projection envelope: %w", err)
	}
	if err := observation.validate(projection.Document); err != nil {
		return result, fmt.Errorf("typed observation envelope: %w", err)
	}
	base, err := Get(projection.Document)
	if err != nil {
		return result, fmt.Errorf("typed projection document: %w", err)
	}
	if !SemanticEquivalent(base, projection.Base) {
		return result, errors.New("typed projection base model is stale")
	}
	accepted, err := Reconcile(base, projection.AcceptedDelta)
	if err != nil {
		return result, fmt.Errorf("typed projection accepted delta: %w", err)
	}
	if err := CheckPutGet(projection.Document, accepted.Model); err != nil {
		return result, fmt.Errorf("typed projection Put-Get law: %w", err)
	}
	threeWay, err := ReconcileThreeWay(projection.Base, projection.Left, projection.Right)
	if err != nil {
		return result, fmt.Errorf("typed projection three-way law: %w", err)
	}
	if !SemanticEquivalent(accepted.Model, threeWay.Model) {
		return result, errors.New("typed projection three-way result diverges from accepted model")
	}
	fixture := bxTypedEnvelopeFixture{projection: projection, observation: observation}
	evidence, err := MeasureBXFixture(fixture)
	if err != nil {
		return result, fmt.Errorf("typed projection evidence: %w", err)
	}
	result = typedEnvelopeResult(projection, observation, threeWay, evidence)
	result.CanonicalJSON, err = typedEnvelopeJSON(result)
	if err != nil {
		return BXTypedEnvelopeResult{}, fmt.Errorf("typed projection canonical evidence: %w", err)
	}
	result.Hash = digest(result.CanonicalJSON)
	if err := result.Validate(); err != nil {
		return BXTypedEnvelopeResult{}, err
	}
	return result, nil
}
