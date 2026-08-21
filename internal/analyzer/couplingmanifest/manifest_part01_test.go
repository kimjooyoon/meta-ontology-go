package couplingmanifest

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	detector "github.com/kimjooyoon/meta-ontology-go/internal/detection/coupling"
	"testing"
)

var (
	_ Manifest                     = detector.ChangeManifest{}
	_ Result                       = detector.Result{}
	_ func([]byte) (Result, error) = DecodeResult
)

func TestBuildUsesDetectorAuthorityAndPassThroughCodecs(t *testing.T) {
	source := testSource(t, "pkg/order.go", "Order", "urn:gooo:entity:order")
	input := testInput(t, []selectiveci.SourceInput{source}, []selectiveci.SourceInput{source}, []surfaceFixture{{Owner: source.Bindings[0].ID, Suffix: "order"}})
	output, err := BuildDetailed(input)
	if err != nil {
		t.Fatalf("BuildDetailed: %v", err)
	}
	if !output.Manifest.Complete || !output.Manifest.ZeroChange || output.Manifest.Schema != detector.ManifestSchemaV1 {
		t.Fatalf("manifest = %#v", output.Manifest)
	}
	if output.DetectorResult.Status != detector.StatusUnknown || output.DetectorResult.Reasons[0].Code != detector.ReasonExternalReceiptMissing {
		t.Fatalf("structural detector result = %#v", output.DetectorResult)
	}
	if output.Metadata.Status != ConstructionComplete || output.Metadata.Counts != (ComponentCounts{Registered: 1, Before: 1, Head: 1, Resolved: 1}) {
		t.Fatalf("metadata = %#v", output.Metadata)
	}

	packet := detectorInput(output.Manifest, input.Authority)
	wantBytes, err := detector.EncodeInput(packet)
	if err != nil {
		t.Fatalf("detector EncodeInput: %v", err)
	}
	gotBytes, err := EncodeInput(packet)
	if err != nil || !bytes.Equal(gotBytes, wantBytes) {
		t.Fatalf("adapter changed detector input bytes: %v", err)
	}
	decoded, err := DecodeJSON(wantBytes)
	if err != nil {
		t.Fatalf("adapter DecodeJSON: %v", err)
	}
	replayed, err := detector.EncodeInput(decoded)
	if err != nil || !bytes.Equal(replayed, wantBytes) {
		t.Fatalf("detector input replay changed bytes: %v", err)
	}
	resultBytes, err := detector.EncodeResult(output.DetectorResult)
	if err != nil {
		t.Fatalf("detector EncodeResult: %v", err)
	}
	decodedResult, err := DecodeResult(resultBytes)
	if err != nil {
		t.Fatalf("adapter DecodeResult: %v", err)
	}
	replayedResult, err := EncodeResult(decodedResult)
	if err != nil || !bytes.Equal(replayedResult, resultBytes) {
		t.Fatalf("detector result replay changed bytes: %v", err)
	}
}
