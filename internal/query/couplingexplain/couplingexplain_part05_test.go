package couplingexplain

import (
	"bytes"
	"context"
	"testing"
)

func TestCancellationVersionRaceAndNoWrite(t *testing.T) {
	request, envelope := fixtureEnvelope(t, ClaimDelta, VerdictVerified)
	original := append([]byte(nil), mustJSON(t, envelope)...)
	raceRequest := request
	raceRequest.Control.ObservedVersion++
	got := Explain(context.Background(), raceRequest, envelope)
	if got.Status != StatusUnknown || got.NoLink == nil || got.NoLink.Reason != ReasonStale || got.Link != nil {
		t.Fatalf("version race = %#v", got)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got = Explain(ctx, request, envelope)
	if got.Status != StatusUnknown || got.NoLink == nil || got.NoLink.Reason != ReasonStale || got.Link != nil {
		t.Fatalf("cancellation = %#v", got)
	}
	if !bytes.Equal(original, mustJSON(t, envelope)) {
		t.Fatal("Explain mutated its envelope input")
	}
}
