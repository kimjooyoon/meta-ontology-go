package generator

import (
	"bytes"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"strings"
	"testing"
)

func TestEntityFieldsSupportedProjectionPreservesOrderIdentityAndMetadata(t *testing.T) {
	ir := entityFieldsFixture()
	result := supportedEntityFieldsResult(t, ir, nil)
	source := string(result.Source)
	first := strings.Index(source, "OrderNumber")
	second := strings.Index(source, "CustomerName")
	if first < 0 || second < 0 || first >= second {
		t.Fatalf("field order or type was not projected exactly: %s", source)
	}
	if strings.Contains(source, "orderNumber") || strings.Contains(source, "customerName") {
		t.Fatalf("field names were transformed instead of using Field.Name: %s", source)
	}
	markers, err := parseMarkers(result.Source)
	if err != nil || len(markers.Regions) != 2 {
		t.Fatalf("generated marker manifest is incomplete: %#v %v", markers, err)
	}
	fieldMappings := make([]SourceMapping, 0, 2)
	for _, mapping := range result.SourceMap.Mappings {
		if mapping.Kind == "field" {
			fieldMappings = append(fieldMappings, mapping)
		}
	}
	if len(fieldMappings) != 2 {
		t.Fatalf("field source-map totality failed: %#v", result.SourceMap)
	}
	fixture := entityFieldsFixture()
	for index, mapping := range fieldMappings {
		field := fixture.Entities[0].Fields[index]
		if mapping.SemanticID != field.ID || mapping.Source != field.Source || mapping.NameSource != field.NameSpan || mapping.ParentID != field.Parent || mapping.TypeRefID != field.TypeRefID || mapping.Presence != field.Presence || mapping.Cardinality != field.Cardinality {
			t.Fatalf("field %d lost authoritative metadata: %#v", index, mapping)
		}
		if mapping.ProfileID != syntax.EntityFieldsProfileID || mapping.ProfileVersion != syntax.EntityFieldsProfileVersion || mapping.ProfileDigest != syntax.EntityFieldsProfileDigest {
			t.Fatalf("field %d lost profile binding: %#v", index, mapping)
		}
		if mapping.Generated.Start.Offset >= mapping.Generated.End.Offset {
			t.Fatalf("field %d has empty generated range: %#v", index, mapping)
		}
	}
	if fieldMappings[0].Generated.End.Offset > fieldMappings[1].Generated.Start.Offset {
		t.Fatalf("field generated ranges overlap: %#v", fieldMappings)
	}
	projection, err := generateProjectionV1WithEntityFieldsSupport(New(Options{}), ir, nil, supportedEntityFieldsForTest())
	if err != nil || projection.Metadata.EntityFields == nil || projection.Metadata.EntityFields.State != string(syntax.EntityFieldsSupported) {
		t.Fatalf("supported projection metadata lost its exact state/profile: %#v %v", projection.Metadata, err)
	}
	if projection.Metadata.EntityFields.Profile.ID != syntax.EntityFieldsProfileID || projection.Metadata.EntityFields.Profile.Version != syntax.EntityFieldsProfileVersion || projection.Metadata.EntityFields.Profile.Digest != syntax.EntityFieldsProfileDigest {
		t.Fatalf("supported projection metadata lost its exact profile: %#v", projection.Metadata.EntityFields)
	}
}
func TestEntityFieldsSupportedReplayPreservesHandwrittenSlotBytes(t *testing.T) {
	ir := entityFieldsFixture()
	first := supportedEntityFieldsResult(t, ir, nil)
	previous := bytes.Replace(first.Source, []byte("return Order{}"), []byte("return Order{}\n\t// handwritten bytes  \n"), 1)
	if bytes.Equal(previous, first.Source) {
		t.Fatal("fixture did not produce a replay input")
	}
	replayed := supportedEntityFieldsResult(t, ir, previous)
	if !bytes.Contains(replayed.Source, []byte("// handwritten bytes  ")) {
		t.Fatalf("handwritten slot bytes were not preserved: %s", replayed.Source)
	}
	if len(replayed.SourceMap.Lookup("urn:gooo:field:order-number")) != 1 || len(replayed.SourceMap.Lookup("urn:gooo:slot:load-order")) != 1 {
		t.Fatalf("replay lost field or slot mappings: %#v", replayed.SourceMap)
	}
}
