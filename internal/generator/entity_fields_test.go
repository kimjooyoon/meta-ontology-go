package generator

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func supportedEntityFieldsForTest() entityFieldsSupport {
	support := checkedEntityFieldsSupport()
	support.State = syntax.EntityFieldsSupported
	return support
}

func entityFieldsFixture() SemanticIR {
	return SemanticIR{
		Package: "entityfieldsgen",
		Entities: []Entity{{
			ID: "urn:gooo:entity:order", Name: "Order", GoName: "Order",
			Source: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 100, Line: 8, Column: 1}},
			Fields: []Field{
				{
					ID: "urn:gooo:field:order-number", Parent: "urn:gooo:entity:order", Name: "OrderNumber",
					TypeRefID: entityFieldsStringTypeID, Presence: "required", Cardinality: "one",
					Source:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 20, Line: 4, Column: 5}, End: Position{Offset: 38, Line: 4, Column: 23}},
					IDSpan:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 20, Line: 4, Column: 5}, End: Position{Offset: 22, Line: 4, Column: 7}},
					NameSpan:        SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 23, Line: 4, Column: 8}, End: Position{Offset: 25, Line: 4, Column: 10}},
					TypeRefSpan:     SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 26, Line: 4, Column: 11}, End: Position{Offset: 28, Line: 4, Column: 13}},
					PresenceSpan:    SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 29, Line: 4, Column: 14}, End: Position{Offset: 32, Line: 4, Column: 17}},
					CardinalitySpan: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 33, Line: 4, Column: 18}, End: Position{Offset: 37, Line: 4, Column: 22}},
				},
				{
					ID: "urn:gooo:field:customer-name", Parent: "urn:gooo:entity:order", Name: "CustomerName",
					TypeRefID: entityFieldsStringTypeID, Presence: "required", Cardinality: "one",
					Source:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 40, Line: 5, Column: 5}, End: Position{Offset: 58, Line: 5, Column: 23}},
					IDSpan:          SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 40, Line: 5, Column: 5}, End: Position{Offset: 42, Line: 5, Column: 7}},
					NameSpan:        SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 43, Line: 5, Column: 8}, End: Position{Offset: 45, Line: 5, Column: 10}},
					TypeRefSpan:     SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 46, Line: 5, Column: 11}, End: Position{Offset: 48, Line: 5, Column: 13}},
					PresenceSpan:    SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 49, Line: 5, Column: 14}, End: Position{Offset: 52, Line: 5, Column: 17}},
					CardinalitySpan: SourceSpan{URI: "entity-fields.gooo", Start: Position{Offset: 53, Line: 5, Column: 18}, End: Position{Offset: 57, Line: 5, Column: 22}},
				},
			},
		}},
		Activities: []Activity{{
			ID: "urn:gooo:activity:load-order", Name: "LoadOrder", GoName: "LoadOrder",
			Outputs: []Port{{ID: "urn:gooo:entity:order", EntityID: "urn:gooo:entity:order", Name: "order", GoName: "order", GoType: "Order"}},
			Slots:   []Slot{{ID: "urn:gooo:slot:load-order", Default: "return Order{}"}},
		}},
	}
}

func supportedEntityFieldsResult(t *testing.T, ir SemanticIR, previous []byte) Result {
	t.Helper()
	result, err := New(Options{}).generateWithEntityFieldsSupport(ir, previous, supportedEntityFieldsForTest())
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestEntityFieldsProductionEntryPointsRemainDeferred(t *testing.T) {
	ir := entityFieldsFixture()
	cases := []struct {
		name string
		call func() (Result, error)
	}{
		{name: "project", call: func() (Result, error) { return Project(ir, nil) }},
		{name: "generate", call: func() (Result, error) { return Generate(ir, nil) }},
		{name: "generator method", call: func() (Result, error) { return New(Options{}).Generate(ir, nil) }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := testCase.call()
			assertEntityFieldsDeferred(t, err)
			if result.Source != nil || result.SourceMap.Mappings != nil {
				t.Fatalf("deferred production entry point returned artifacts: %#v", result)
			}
		})
	}

	t.Run("generate from", func(t *testing.T) {
		source, sourceMap, err := GenerateFrom(ir, Options{})
		assertEntityFieldsDeferred(t, err)
		if source != nil || sourceMap.Mappings != nil {
			t.Fatalf("GenerateFrom returned artifacts: %q %#v", source, sourceMap)
		}
	})
	t.Run("projection metadata", func(t *testing.T) {
		result, err := GenerateProjectionV1(ir, nil)
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateProjectionV1 returned artifacts: %#v", result)
		}
	})
	t.Run("adapter projection metadata", func(t *testing.T) {
		result, err := GenerateFromProjectionV1(ir, Options{})
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateFromProjectionV1 returned artifacts: %#v", result)
		}
	})
	t.Run("metadata", func(t *testing.T) {
		result, err := GenerateWithMetadata(ir, nil)
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Metadata.SourceDigest != "" {
			t.Fatalf("GenerateWithMetadata returned artifacts: %#v", result)
		}
	})
	t.Run("binding", func(t *testing.T) {
		result, err := GenerateWithBinding(ir, nil, ProjectionBinding{})
		assertEntityFieldsDeferred(t, err)
		if result.Source != nil || result.SourceMap.Mappings != nil || result.Schema != "" {
			t.Fatalf("GenerateWithBinding returned artifacts: %#v", result)
		}
	})
}

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

func TestEntityFieldsProfileBindingFailsClosed(t *testing.T) {
	ir := entityFieldsFixture()
	cases := []struct {
		name string
		edit func(*entityFieldsSupport)
		code string
	}{
		{name: "unknown state", edit: func(s *entityFieldsSupport) { s.State = "UNKNOWN" }, code: entityFieldsUnknownStateDiagnostic},
		{name: "unbound", edit: func(s *entityFieldsSupport) { s.Profile = syntax.EntityFieldsProfile{} }, code: entityFieldsUnboundProfileDiagnostic},
		{name: "profile ID", edit: func(s *entityFieldsSupport) { s.Profile.ID = "other" }, code: entityFieldsProfileMismatchDiagnostic},
		{name: "profile version", edit: func(s *entityFieldsSupport) { s.Profile.Version++ }, code: entityFieldsProfileMismatchDiagnostic},
		{name: "profile digest", edit: func(s *entityFieldsSupport) { s.Profile.Digest = "tampered" }, code: entityFieldsProfileDigestDiagnostic},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			support := supportedEntityFieldsForTest()
			testCase.edit(&support)
			result, err := New(Options{}).generateWithEntityFieldsSupport(ir, nil, support)
			if err == nil || !strings.Contains(err.Error(), testCase.code) || result.Source != nil || result.SourceMap.Mappings != nil {
				t.Fatalf("profile rejection was not transactional: result=%#v err=%v", result, err)
			}
		})
	}
}

func TestEntityFieldsFailuresDoNotMutateCallerInputsOrPreviousSource(t *testing.T) {
	ir := entityFieldsFixture()
	first := supportedEntityFieldsResult(t, ir, nil)
	previous := append([]byte(nil), first.Source...)
	changed := copyIR(ir)
	changed.Entities[0].Fields[0].Presence = "optional"
	beforeIR := copyIR(changed)
	result, err := New(Options{}).generateWithEntityFieldsSupport(changed, previous, supportedEntityFieldsForTest())
	if err == nil || result.Source != nil || !bytes.Equal(previous, first.Source) || !reflect.DeepEqual(changed, beforeIR) {
		t.Fatalf("rejected generation was not no-write: result=%#v err=%v", result, err)
	}
	corrupt := append(append([]byte(nil), first.Source...), []byte("\n//gooo:slot:start id=\"orphan\"\n")...)
	beforeCorrupt := append([]byte(nil), corrupt...)
	result, err = New(Options{}).generateWithEntityFieldsSupport(ir, corrupt, supportedEntityFieldsForTest())
	if err == nil || result.Source != nil || !bytes.Equal(corrupt, beforeCorrupt) {
		t.Fatalf("tampered previous source was not no-write: result=%#v err=%v", result, err)
	}
}

func TestFieldlessProductionProjectionKeepsDeferredMetadataAbsentAndDeterministic(t *testing.T) {
	ir := SemanticIR{Package: "fieldless", Entities: []Entity{{ID: "urn:gooo:entity:empty", Name: "Empty", GoName: "Empty"}}}
	first, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateProjectionV1(ir, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.Source, second.Source) || !reflect.DeepEqual(first.SourceMap, second.SourceMap) || first.Metadata.EntityFields != nil || first.Metadata.SourceDigest != second.Metadata.SourceDigest || first.Metadata.SemanticIRDigest != second.Metadata.SemanticIRDigest || first.Metadata.SourceMapDigest != second.Metadata.SourceMapDigest {
		t.Fatalf("fieldless projection was not byte/hash deterministic: first=%#v second=%#v", first.Metadata, second.Metadata)
	}
	canonical, err := first.CanonicalJSON()
	if err != nil || bytes.Contains(canonical, []byte("entity_fields")) {
		t.Fatalf("fieldless canonical metadata changed: err=%v json=%s", err, canonical)
	}
}

func assertEntityFieldsDeferred(t *testing.T, err error) {
	t.Helper()
	if err == nil || !strings.Contains(err.Error(), entityFieldsDeferredDiagnostic) {
		t.Fatalf("expected deterministic deferred error, got %v", err)
	}
}
