package bidir

import (
	"fmt"
)

type bxReceiptObserver struct{ receipt BXObserverReceipt }

func (observer *bxReceiptObserver) Kind() string { return observer.receipt.ObserverKind }
func (observer *bxReceiptObserver) ObserveRejected(operation func() error) (BXWriteObservation, error) {
	if operation == nil {
		return BXWriteObservation{}, fmt.Errorf("rejected operation is nil")
	}
	_ = operation()
	return cloneWriteObservation(observer.receipt.Observation), nil
}
func cloneWriteObservation(observation BXWriteObservation) BXWriteObservation {
	return BXWriteObservation{Observed: observation.Observed, Before: cloneSnapshot(observation.Before), After: cloneSnapshot(observation.After)}
}
func typedEnvelopeJSON(result BXTypedEnvelopeResult) (string, error) {
	value := struct {
		SchemaVersion string   `json:"schema_version"`
		Fixture       string   `json:"fixture"`
		FeatureGreen  bool     `json:"feature_green"`
		Projection    string   `json:"projection_digest"`
		Base          string   `json:"base_fingerprint"`
		Left          string   `json:"left_fingerprint"`
		Right         string   `json:"right_fingerprint"`
		GetPut        bool     `json:"get_put"`
		PutGet        bool     `json:"put_get"`
		Equivalent    bool     `json:"semantic_equivalence"`
		ThreeWay      string   `json:"three_way_fingerprint"`
		Delta         string   `json:"delta_json"`
		Partial       string   `json:"partial_delta_json"`
		Locality      string   `json:"locality_json"`
		Candidates    []string `json:"candidates"`
		Ports         []string `json:"port_sequence"`
		Relations     []string `json:"relation_sequence"`
		Observer      string   `json:"observer"`
		NoWrite       bool     `json:"no_write"`
		Deferred      []string `json:"deferred"`
	}{
		SchemaVersion: result.SchemaVersion, Fixture: result.Fixture, FeatureGreen: result.FeatureGreen,
		Projection: result.ProjectionDigest, Base: result.BaseFingerprint, Left: result.LeftFingerprint, Right: result.RightFingerprint,
		GetPut: result.GetPutPassed, PutGet: result.PutGetPassed, Equivalent: result.SemanticEquivalent,
		ThreeWay: SemanticFingerprint(result.ThreeWay.Model), Delta: result.Delta.CanonicalJSON,
		Partial: result.PartialDelta.CanonicalJSON, Locality: localityJSON(result.Locality, localityDigest(result.Locality)),
		Candidates: result.Candidates, Ports: result.PortSequence, Relations: result.RelationSequence,
		Observer: result.ObserverKind, NoWrite: result.NoWriteObserved, Deferred: result.Deferred,
	}
	return canonicalJSON(value)
}
