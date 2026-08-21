package coupling

import (
	"encoding/json"
	"testing"
)

func marshalWithoutRebindingEvidence(t *testing.T, envelope Envelope) []byte {
	t.Helper()
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func TestProtocolTranscriptReadOnlyStandardSurface(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	wire, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(wire, &fields); err != nil {
		t.Fatal(err)
	}
	for field := range fields {
		switch field {
		case "Outcome", "Links", "Hover", "Diagnostics":
		default:
			t.Fatalf("read-only adapter exposed an unexpected wire field %q: %s", field, wire)
		}
	}
	if len(result.Links) != 1 || len(result.Diagnostics) != 1 {
		t.Fatalf("read-only result = %#v", result)
	}
}
