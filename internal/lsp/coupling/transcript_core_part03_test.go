package coupling

import (
	"context"
	"testing"
)

func TestProtocolTranscriptCancellationWinsOrdering(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request := transcriptRequest(envelope)
	request.Context = ctx
	request.DocumentVersion++
	request.SnapshotDigest = transcriptDigest("stale")
	result := transcriptAdapter(t, envelope).Resolve(request)
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Hover != nil {
		t.Fatalf("cancelled result = %#v", result)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != DiagnosticCancelled {
		t.Fatalf("cancelled diagnostics = %#v", result.Diagnostics)
	}
}
