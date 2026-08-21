package bidir

import (
	"testing"
)

func typedEnvelopeFixture(t *testing.T) (BXTypedProjectionEnvelope, BXTypedObservationEnvelope) {
	t.Helper()
	fixture := billingBXFixture{}
	document := fixture.Document()
	base, err := Get(document)
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := Reconcile(base, fixture.AcceptedDelta())
	if err != nil {
		t.Fatal(err)
	}
	updated, err := Put(document, accepted.Model)
	if err != nil {
		t.Fatal(err)
	}
	observer := NewBXMemoryRejectedWriteObserver(document)
	receipt, err := CaptureBXObserverReceipt(observer, func() error {
		_, reconcileErr := Reconcile(base, fixture.PartialDelta())
		return reconcileErr
	})
	if err != nil {
		t.Fatal(err)
	}
	projection := BXTypedProjectionEnvelope{
		SchemaVersion: BXTypedEnvelopeSchemaVersion, Fixture: "typed-billing", Document: document,
		Base: base, Left: accepted.Model, Right: base.Clone(), AcceptedDelta: fixture.AcceptedDelta(),
		PartialDelta: fixture.PartialDelta(), BaseEvidence: fixture.BaseEvidence(),
	}
	projection, err = SealBXTypedProjectionEnvelope(projection)
	if err != nil {
		t.Fatal(err)
	}
	observation := NewBXTypedObservationEnvelope(fixtureWriteObservation(document, updated), receipt)
	return projection, observation
}
