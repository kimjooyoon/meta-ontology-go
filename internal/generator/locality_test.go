package generator

import (
	"bytes"
	"go/format"
	"strings"
	"testing"
)

func TestGeneratePreservesUnrelatedRegionsAndMapsLocalChanges(t *testing.T) {
	first, err := Generate(billingIR(), nil)
	if err != nil {
		t.Fatal(err)
	}
	entityBefore := testGeneratedBlock(t, first.Source, "billing://entity/order")
	previous := strings.Replace(string(first.Source), "package billinggen\n", "package billinggen\n\n// handwritten outside the generated regions\nvar Keep = 7\n", 1)

	changed := billingIR()
	changed.Activities[0].Name = "AuthorizeOrder"
	changed.Activities[0].GoName = "AuthorizeOrder"
	second, err := Generate(changed, []byte(previous))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(second.Source), "// handwritten outside the generated regions\nvar Keep = 7") {
		t.Fatalf("marker-outside handwritten text was changed:\n%s", second.Source)
	}
	if !bytes.Equal(entityBefore, testGeneratedBlock(t, second.Source, "billing://entity/order")) {
		t.Fatalf("unrelated entity region changed")
	}
	if !strings.Contains(string(second.Source), "func AuthorizeOrder(") || strings.Contains(string(second.Source), "func PayOrder(") {
		t.Fatalf("changed activity region was not localized:\n%s", second.Source)
	}
	formatted, err := format.Source(second.Source)
	if err != nil {
		t.Fatalf("output is not gofmt-compatible: %v", err)
	}
	if !bytes.Equal(formatted, second.Source) {
		t.Fatalf("output is not gofmt-stable:\n%s", second.Source)
	}

	activityMappings := second.SourceMap.Lookup("billing://activity/pay-order")
	if len(activityMappings) != 1 || len(second.SourceMap.Lookup("billing://activity/pay-order/implementation")) != 1 {
		t.Fatalf("expected activity and implementation-slot mappings, got %#v", second.SourceMap)
	}
	for _, mapping := range activityMappings {
		if mapping.Generated.Start.Offset < 0 || mapping.Generated.End.Offset > len(second.Source) || mapping.Generated.Start.Offset > mapping.Generated.End.Offset {
			t.Fatalf("invalid generated range: %#v", mapping)
		}
	}
}

func TestGenerateRejectsMalformedMarkers(t *testing.T) {
	_, err := Generate(billingIR(), []byte("package billinggen\n\n//gooo:generated:start id=\"orphan\" kind=\"activity\"\n"))
	if err == nil || !strings.Contains(err.Error(), "unterminated generated region") {
		t.Fatalf("expected malformed marker error, got %v", err)
	}
}

func testGeneratedBlock(t *testing.T, source []byte, id string) []byte {
	t.Helper()
	markers, err := parseMarkers(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, region := range markers.Regions {
		if region.ID == id {
			return append([]byte(nil), source[region.Start:region.End]...)
		}
	}
	t.Fatalf("generated region %q not found", id)
	return nil
}
