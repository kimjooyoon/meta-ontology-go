package coupling

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
	"testing"
)

func TestLiveQueryReplayLabelAndPresentationStability(t *testing.T) {
	adapter := liveAdapter(t)
	request := liveRequest()
	first := adapter.Resolve(request)
	second := adapter.Resolve(request)
	if first.Outcome != OutcomePass || second.Outcome != OutcomePass {
		t.Fatalf("replay result = %#v / %#v", first, second)
	}
	firstJSON, err := json.Marshal(first.Diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second.Diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("replay diagnostics changed: %s / %s", firstJSON, secondJSON)
	}
	firstWire, err := json.Marshal(struct {
		Links       []LocationLink `json:"links"`
		Hover       *Hover         `json:"hover"`
		Diagnostics []Diagnostic   `json:"diagnostics"`
	}{first.Links, first.Hover, first.Diagnostics})
	if err != nil {
		t.Fatal(err)
	}
	secondWire, err := json.Marshal(struct {
		Links       []LocationLink `json:"links"`
		Hover       *Hover         `json:"hover"`
		Diagnostics []Diagnostic   `json:"diagnostics"`
	}{second.Links, second.Hover, second.Diagnostics})
	if err != nil {
		t.Fatal(err)
	}
	if string(firstWire) != string(secondWire) {
		t.Fatalf("replay standard output changed: %s / %s", firstWire, secondWire)
	}
	envelope, err := couplingexplain.DecodeVerifiedEnvelope([]byte(literalVerifiedQueryEnvelope))
	if err != nil {
		t.Fatal(err)
	}
	envelope.Term.Presentation.Label = "Renamed presentation"
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	renamed := ResolveLive(request, data)
	if renamed.Outcome != OutcomePass || len(renamed.Links) != 1 || renamed.Diagnostics[0].Code != DiagnosticExplanation {
		t.Fatalf("presentation mutation changed authority decision: %#v", renamed)
	}
}
