package bidir

import (
	"bytes"
	"fmt"
	"strings"
)

func typedProjectionDigest(envelope BXTypedProjectionEnvelope) string {
	parts := []string{
		envelope.SchemaVersion, envelope.Fixture, documentDigest(envelope.Document),
		SemanticFingerprint(envelope.Base), SemanticFingerprint(envelope.Left), SemanticFingerprint(envelope.Right),
		factSequenceHash(envelope.AcceptedDelta), factOrderHash(envelope.AcceptedDelta),
		factSequenceHash(envelope.PartialDelta), factOrderHash(envelope.PartialDelta),
		typedBaseEvidenceDigest(envelope.BaseEvidence),
	}
	return digest(strings.Join(parts, "|"))
}
func typedBaseEvidenceDigest(input BXBaseEvidenceInput) string {
	parts := []string{
		documentDigest(input.DSL), SemanticFingerprint(input.IR), digestFacts(input.Go),
		digestSpans(input.SourceMap), digestFacts(input.Evidence), digestSpans(input.Provenance),
	}
	return digest(strings.Join(parts, "|"))
}
func observerReceiptDigest(receipt BXObserverReceipt) string {
	return digest(strings.Join([]string{receipt.ObserverKind, snapshotDigest(receipt.Observation.Before), snapshotDigest(receipt.Observation.After), fmt.Sprintf("%t", receipt.Observation.Observed)}, "|"))
}
func snapshotDigest(snapshot BXFileSnapshot) string {
	return digest(fmt.Sprintf("%s|%d|%d|%t|%s", snapshot.LStat.Path, snapshot.LStat.Size, snapshot.LStat.Mode, snapshot.LStat.Exists, snapshot.Bytes))
}
func sameSnapshot(left, right BXFileSnapshot) bool {
	return bytes.Equal(left.Bytes, right.Bytes) && left.LStat == right.LStat
}
func typedEnvelopeResult(projection BXTypedProjectionEnvelope, observation BXTypedObservationEnvelope, threeWay ThreeWayResult, evidence BXEvidence) BXTypedEnvelopeResult {
	return BXTypedEnvelopeResult{
		SchemaVersion: BXTypedEnvelopeSchemaVersion, Fixture: projection.Fixture, FeatureGreen: false,
		ProjectionDigest: projection.Digest, BaseFingerprint: SemanticFingerprint(projection.Base),
		LeftFingerprint: SemanticFingerprint(projection.Left), RightFingerprint: SemanticFingerprint(projection.Right),
		GetPutPassed: evidence.GetPutPassed, PutGetPassed: true, SemanticEquivalent: evidence.SemanticEquivalent,
		ThreeWay: threeWay, Evidence: evidence, Delta: evidence.Delta, PartialDelta: evidence.PartialDelta,
		Locality: detachedLocality(evidence.Locality), Candidates: append([]string{}, evidence.Delta.Candidates...),
		PortSequence: append([]string{}, evidence.Delta.PortSequence...), RelationSequence: append([]string{}, evidence.Delta.RelationSequence...),
		ObserverKind: observation.Rejected.ObserverKind, NoWriteObserved: evidence.RejectedTransaction.NoWrite,
		Deferred: append([]string{}, evidence.Deferred...),
	}
}

type bxTypedEnvelopeFixture struct {
	projection  BXTypedProjectionEnvelope
	observation BXTypedObservationEnvelope
}

func (fixture bxTypedEnvelopeFixture) Name() string       { return fixture.projection.Fixture }
func (fixture bxTypedEnvelopeFixture) Document() Document { return fixture.projection.Document }
func (fixture bxTypedEnvelopeFixture) AcceptedDelta() FactDelta {
	return fixture.projection.AcceptedDelta
}
func (fixture bxTypedEnvelopeFixture) PartialDelta() FactDelta {
	return fixture.projection.PartialDelta
}
func (fixture bxTypedEnvelopeFixture) BaseEvidence() BXBaseEvidenceInput {
	return fixture.projection.BaseEvidence
}
func (fixture bxTypedEnvelopeFixture) ObserveAcceptedWrite(before, after Document) BXWriteObservation {
	return fixture.observation.Accepted
}
func (fixture bxTypedEnvelopeFixture) RejectedWriteObserver(document Document) (BXRejectedWriteObserver, error) {
	if fixture.observation.Rejected == nil {
		return nil, fmt.Errorf("observer receipt is absent")
	}
	return &bxReceiptObserver{receipt: *fixture.observation.Rejected}, nil
}
