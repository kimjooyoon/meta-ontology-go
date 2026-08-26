package coupling

import (
	"context"
	"testing"
)

func TestLiveQueryNilAndMalformedGuards(t *testing.T) {
	var nilAdapter *LiveAdapter
	request := liveRequest()
	result := nilAdapter.Resolve(request)
	if len(result.Links) != 0 || result.Outcome != OutcomeFailClosed {
		t.Fatalf("nil adapter = %#v", result)
	}
	result = ResolveLive(request, nil)
	if len(result.Links) != 0 || result.Outcome != OutcomeFailClosed {
		t.Fatalf("nil input = %#v", result)
	}
	result = ResolveLive(request, append([]byte(literalVerifiedQueryEnvelope), []byte(` {"trailing":true}`)...))
	if len(result.Links) != 0 || result.Outcome != OutcomeFailClosed {
		t.Fatalf("trailing input = %#v", result)
	}
}
func TestLiveQueryCancellationPrecedesMalformedEnvelope(t *testing.T) {
	request := liveRequest()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	request.Context = ctx
	result := ResolveLive(request, []byte("{"))
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticCancelled {
		t.Fatalf("cancelled malformed result = %#v", result)
	}

	request = liveRequest()
	request.Query.Control.ObservedCancellationVersion = request.Query.Control.RequestCancellationVersion - 1
	result = ResolveLive(request, []byte("{"))
	if result.Outcome != OutcomeUnknown || len(result.Links) != 0 || result.Diagnostics[0].Code != DiagnosticCancelled {
		t.Fatalf("out-of-order cancellation result = %#v", result)
	}
}
func TestLiveUTF16PositionConversion(t *testing.T) {
	source := "😀 cafe\r\n두 번째"
	position, err := OffsetToPosition(source, len("😀 "))
	if err != nil || position != (Position{Line: 0, Character: 3}) {
		t.Fatalf("UTF-16 position = %#v, err=%v", position, err)
	}
	offset, err := PositionToOffset(source, position)
	if err != nil || offset != len("😀 ") {
		t.Fatalf("UTF-16 offset = %d, err=%v", offset, err)
	}
	if _, err := PositionToOffset(source, Position{Line: 0, Character: 1}); err == nil {
		t.Fatal("accepted astral-code-point split")
	}
	if _, err := OffsetToPosition(source, len("😀")-1); err == nil {
		t.Fatal("accepted UTF-8 byte split")
	}
	if _, _, err := ValidateRange(source, Range{Start: position, End: Position{Line: 0, Character: 1}}); err == nil {
		t.Fatal("accepted reversed UTF-16 range")
	}
}
