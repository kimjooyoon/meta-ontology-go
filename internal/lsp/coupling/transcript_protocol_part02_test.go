package coupling

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestProtocolTranscriptStandardClientJSONRoundTrip(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	result := transcriptAdapter(t, envelope).Resolve(transcriptRequest(envelope))
	linksJSON, err := json.Marshal(result.Links)
	if err != nil {
		t.Fatal(err)
	}
	var links []LocationLink
	if err := json.Unmarshal(linksJSON, &links); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Links, links) {
		t.Fatalf("location links changed after JSON roundtrip: %#v %#v", result.Links, links)
	}
	allJSON, err := json.Marshal(struct {
		Links       []LocationLink `json:"links"`
		Hover       *Hover         `json:"hover"`
		Diagnostics []Diagnostic   `json:"diagnostics"`
	}{result.Links, result.Hover, result.Diagnostics})
	if err != nil {
		t.Fatal(err)
	}
	wire := string(allJSON)
	for _, forbidden := range []string{"stable_id", "code_symbol_id", "semantic_owner_id", "customWireID"} {
		if strings.Contains(wire, forbidden) {
			t.Fatalf("standard LSP wire contains custom identity field %q: %s", forbidden, wire)
		}
	}
	if !strings.Contains(wire, "relatedInformation") || !strings.Contains(wire, "targetUri") {
		t.Fatalf("standard LSP fields missing from wire: %s", wire)
	}
}
func TestProtocolTranscriptZeroWritesAndInputImmutability(t *testing.T) {
	envelope := transcriptEnvelope("file:///workspace/main.go", "PayOrder", ClaimDelta)
	input := transcriptBytes(t, envelope)
	original := append([]byte(nil), input...)
	adapter, err := New(input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] ^= 1
	if result := adapter.Resolve(transcriptRequest(envelope)); result.Outcome != OutcomePass {
		t.Fatalf("input mutation changed result: %#v", result)
	}
	if !bytes.Equal(adapter.RawBytes(), original) {
		t.Fatal("adapter retained caller-owned mutable bytes")
	}
	if bytes.Equal(input, original) {
		t.Fatal("test did not mutate caller input")
	}
}
