package coupling

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/query/couplingexplain"
)

func TestLiveQueryLiteralRootRelocationAndStandardRoundTrip(t *testing.T) {
	envelope, err := couplingexplain.DecodeVerifiedEnvelope([]byte(literalVerifiedQueryEnvelope))
	if err != nil {
		t.Fatal(err)
	}
	envelope.CodeBinding.Presentation.Root = "relocated-root"
	envelope.Term.Presentation.Root = "relocated-root"
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	request := liveRequest()
	request.DocumentURI = "file:///relocated-root/billing.gooo"
	request.Locations.DocumentURI = request.DocumentURI
	for index := range request.Locations.Locations {
		request.Locations.Locations[index].URI = strings.Replace(request.Locations.Locations[index].URI, "file:///workspace/", "file:///relocated-root/", 1)
	}
	result := ResolveLive(request, data)
	if result.Outcome != OutcomePass || len(result.Links) != 1 || result.Links[0].TargetURI != "file:///relocated-root/model.gooo" {
		t.Fatalf("relocated result = %#v", result)
	}
	linksJSON, err := json.Marshal(result.Links)
	if err != nil {
		t.Fatal(err)
	}
	var links []LocationLink
	if err := json.Unmarshal(linksJSON, &links); err != nil || len(links) != 1 || links[0].TargetURI != result.Links[0].TargetURI {
		t.Fatalf("location-link roundtrip = %#v, err=%v", links, err)
	}
	hoverJSON, err := json.Marshal(result.Hover)
	if err != nil {
		t.Fatal(err)
	}
	var hover Hover
	if err := json.Unmarshal(hoverJSON, &hover); err != nil || hover.Contents.Value != result.Hover.Contents.Value {
		t.Fatalf("hover roundtrip = %#v, err=%v", hover, err)
	}
	diagnosticJSON, err := json.Marshal(result.Diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	var diagnostics []Diagnostic
	if err := json.Unmarshal(diagnosticJSON, &diagnostics); err != nil || len(diagnostics) != 1 || diagnostics[0].Message != result.Diagnostics[0].Message {
		t.Fatalf("diagnostic roundtrip = %#v, err=%v", diagnostics, err)
	}
}

func TestLiveQueryInputAndResultIsolation(t *testing.T) {
	data := []byte(literalVerifiedQueryEnvelope)
	original := append([]byte(nil), data...)
	adapter, err := NewLiveAdapter(data)
	if err != nil {
		t.Fatal(err)
	}
	first := adapter.Resolve(liveRequest())
	data[0] = '['
	first.Links[0].TargetURI = "file:///mutated.gooo"
	first.Diagnostics[0].RelatedInformation = nil
	if string(adapter.RawBytes()) != string(original) {
		t.Fatal("live adapter retained caller-owned mutable bytes")
	}
	second := adapter.Resolve(liveRequest())
	if second.Outcome != OutcomePass || second.Links[0].TargetURI != "file:///workspace/model.gooo" || len(second.Diagnostics[0].RelatedInformation) == 0 {
		t.Fatalf("input/result mutation changed replay: %#v", second)
	}
}
