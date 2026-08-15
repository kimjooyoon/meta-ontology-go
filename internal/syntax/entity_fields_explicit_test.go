package syntax

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func supportedEntityFields() EntityFieldsSupport {
	support := CurrentEntityFieldsSupport()
	support.State = EntityFieldsSupported
	return support
}

func TestSupportedEntityFieldsParsesOrderedFieldsAndExactSpans(t *testing.T) {
	source := "package billing\r\nnamespace billing\r\n\r\n" +
		"entity Order id \"billing://entity/order\" fields {\r\n" +
		"  // source order is authoritative\r\n" +
		"  field 名称 id \"billing://field/one\" type string required one\r\n" +
		"  field 名称 id \"billing://field/two\" type \"gooo:string\" optional many\r\n" +
		"  field Renamed id \"urn:gooo:field:three\" type \"urn:gooo:type:string\" required many\r\n" +
		"}\r\nactivity PayOrder(Order) -> Order\r\n"
	file, diagnostics := ParseFileWithEntityFieldsSupport("fields.gooo", source, supportedEntityFields())
	if len(diagnostics) != 0 || file == nil {
		t.Fatalf("supported parse = %#v, %#v", file, diagnostics)
	}
	entity := file.Declarations[0].(*EntityDecl)
	if !entity.FieldsPresent || len(entity.Fields) != 3 || entity.Span.End.Offset != strings.Index(source, "}\r\nactivity")+1 {
		t.Fatalf("entity block = %#v", entity)
	}
	wantIDs := []string{"billing://field/one", "billing://field/two", "urn:gooo:field:three"}
	wantTypes := []string{"string", "gooo:string", "urn:gooo:type:string"}
	for index, field := range entity.Fields {
		if field.ID != wantIDs[index] || field.TypeRef.Spelling != wantTypes[index] || field.Span.Filename != "fields.gooo" {
			t.Fatalf("field %d values = %#v", index, field)
		}
		if field.Span.IsEmpty() || field.NameSpan.IsEmpty() || field.IDSpan.IsEmpty() || field.TypeRef.Span.IsEmpty() || field.PresenceSpan.IsEmpty() || field.CardinalitySpan.IsEmpty() {
			t.Fatalf("field %d has incomplete spans = %#v", index, field)
		}
		if got := source[field.Span.Start.Offset:field.Span.End.Offset]; !strings.HasPrefix(got, "field ") || !strings.HasSuffix(got, string(field.Cardinality)) {
			t.Fatalf("field %d full span = %q", index, got)
		}
	}
	if got := entity.Fields[0].NameSpan; got.Start.Column != 9 || got.End.Column != 11 || got.End.Offset-got.Start.Offset != len("名称") {
		t.Fatalf("UTF-8 field span = %#v", got)
	}
	if entity.Fields[0].Name != entity.Fields[1].Name || entity.Fields[0].ID == entity.Fields[1].ID {
		t.Fatalf("same-name distinct-ID fields changed: %#v", entity.Fields)
	}
	if entity.Fields[0].Presence != FieldPresenceRequired || entity.Fields[1].Presence != FieldPresenceOptional || entity.Fields[2].Cardinality != FieldCardinalityMany {
		t.Fatalf("field state changed: %#v", entity.Fields)
	}
	formatted, err := FormatWithEntityFieldsSupport(file, supportedEntityFields())
	if err != nil {
		t.Fatal(err)
	}
	want := "package billing\nnamespace billing\n\nentity Order id \"billing://entity/order\" fields {\n" +
		"    field 名称 id \"billing://field/one\" type string required one\n" +
		"    field 名称 id \"billing://field/two\" type \"gooo:string\" optional many\n" +
		"    field Renamed id \"urn:gooo:field:three\" type \"urn:gooo:type:string\" required many\n" +
		"}\nactivity PayOrder(Order) -> Order\n"
	if formatted != want {
		t.Fatalf("supported format = %q, want %q", formatted, want)
	}
	replayed, replayDiagnostics, replayErr := FormatSourceWithEntityFieldsSupport("fields.gooo", source, supportedEntityFields())
	if replayErr != nil || len(replayDiagnostics) != 0 || replayed != formatted {
		t.Fatalf("source replay = %q, %#v, %v", replayed, replayDiagnostics, replayErr)
	}
	second, secondDiagnostics := ParseFileWithEntityFieldsSupport("replay.gooo", formatted, supportedEntityFields())
	secondFormatted, secondErr := FormatWithEntityFieldsSupport(second, supportedEntityFields())
	if secondErr != nil || len(secondDiagnostics) != 0 || secondFormatted != formatted {
		t.Fatalf("canonical replay = %q, %#v, %v", secondFormatted, secondDiagnostics, secondErr)
	}
}

func TestSupportedEntityFieldsPreservesEmptyBlockAndFixedIDRename(t *testing.T) {
	support := supportedEntityFields()
	empty, diagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity Empty id "urn:empty" fields {}`, support)
	if len(diagnostics) != 0 || empty == nil || !empty.Declarations[0].(*EntityDecl).FieldsPresent {
		t.Fatalf("empty block parse = %#v, %#v", empty, diagnostics)
	}
	emptyText, err := FormatWithEntityFieldsSupport(empty, support)
	if err != nil || !strings.Contains(emptyText, `entity Empty id "urn:empty" fields {}`) {
		t.Fatalf("empty block format = %q, %v", emptyText, err)
	}
	first, firstDiagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity E id "urn:e" fields { field Old id "urn:f" type string required one }`, support)
	second, secondDiagnostics := ParseWithEntityFieldsSupport(`package p namespace n entity E id "urn:e" fields { field New id "urn:f" type string required one }`, support)
	if len(firstDiagnostics) != 0 || len(secondDiagnostics) != 0 || first.Declarations[0].(*EntityDecl).Fields[0].ID != second.Declarations[0].(*EntityDecl).Fields[0].ID {
		t.Fatalf("fixed-ID rename changed identity: %#v %#v", first, second)
	}
}

func TestExplicitEntityFieldsRemainDeferredOnOrdinaryPublicPath(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one }`
	file, diagnostics := ParseFile("fields.gooo", source)
	if file != nil || len(diagnostics) != 1 || diagnostics[0].Code != DiagEntityFieldsDeferred {
		t.Fatalf("ordinary parse = %#v, %#v", file, diagnostics)
	}
	formatted, formatDiagnostics, err := FormatSource("fields.gooo", source)
	if formatted != "" || !reflect.DeepEqual(formatDiagnostics, diagnostics) || err == nil {
		t.Fatalf("ordinary format = %q, %#v, %v", formatted, formatDiagnostics, err)
	}
}

func TestEntityFieldsSupportMismatchFailsClosedForParserAndFormatter(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields {}`
	profile := CurrentEntityFieldsSupport().Profile
	cases := []EntityFieldsSupport{
		{State: "UNKNOWN", Profile: profile},
		{State: EntityFieldsSupported, Profile: EntityFieldsProfile{ID: "wrong", Version: 1, Digest: profile.Digest}},
	}
	for _, support := range cases {
		file, diagnostics := ParseWithEntityFieldsSupport(source, support)
		if file != nil || len(diagnostics) != 1 || diagnostics[0].Code != DiagEntityFieldsConfiguration {
			t.Fatalf("mismatch parse = %#v, %#v", file, diagnostics)
		}
		if _, err := FormatWithEntityFieldsSupport(nil, support); !errors.Is(err, ErrEntityFieldsUnknownState) && !errors.Is(err, ErrEntityFieldsProfileMismatch) {
			t.Fatalf("mismatch format error = %v", err)
		}
	}
}

func TestSupportedEntityFieldsMalformedInputPublishesNoPartialFields(t *testing.T) {
	support := supportedEntityFields()
	seeds := []string{
		`package p namespace n entity E id "urn:e" fields field F id "urn:f" type string required one`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one garbage }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one`,
		`package p namespace n entity E id "urn:e" fields { field Good id "urn:good" type string required one field Bad id "urn:bad" type string required }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type @ required one }`,
	}
	for _, source := range seeds {
		first, firstDiagnostics := ParseWithEntityFieldsSupport(source, support)
		second, secondDiagnostics := ParseWithEntityFieldsSupport(source, support)
		if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) || len(firstDiagnostics) == 0 {
			t.Fatalf("malformed replay changed: %#v %#v / %#v %#v", first, firstDiagnostics, second, secondDiagnostics)
		}
		if first != nil && len(first.Declarations) != 0 && first.Declarations[0].(*EntityDecl).Fields != nil {
			t.Fatalf("malformed source published fields: %#v", first.Declarations[0].(*EntityDecl).Fields)
		}
		formatted, formatDiagnostics, err := FormatSourceWithEntityFieldsSupport("bad.gooo", source, support)
		if formatted != "" || len(formatDiagnostics) == 0 || err == nil {
			t.Fatalf("malformed format wrote output: %q %#v %v", formatted, formatDiagnostics, err)
		}
	}
}

func TestSupportedEntityFieldsParserReuseIsDeterministic(t *testing.T) {
	source := `package p namespace n entity E id "urn:e" fields { field F id "urn:f" type "gooo:string" optional many }`
	parser := NewParserWithEntityFieldsSupport(source, supportedEntityFields())
	first, firstDiagnostics := parser.Parse()
	second, secondDiagnostics := parser.Parse()
	if first != second || !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Fatal("explicit parser reuse was not immutable and deterministic")
	}
}

func FuzzParseEntityFieldsSupported(f *testing.F) {
	for _, seed := range []string{
		`package p namespace n entity E id "urn:e" fields {}`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type string required one }`,
		`package p namespace n entity E id "urn:e" fields { field F id "urn:f" type "gooo:string" optional many }`,
		"package p\nnamespace n\nentity E id \"urn:e\" fields { field 名 id \"urn:f\" type \"urn:t\" required one }",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		first, firstDiagnostics := ParseWithEntityFieldsSupport(source, supportedEntityFields())
		second, secondDiagnostics := ParseWithEntityFieldsSupport(source, supportedEntityFields())
		if !reflect.DeepEqual(first, second) || !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
			t.Fatal("explicit parser replay changed")
		}
		if first != nil && !firstDiagnostics.HasErrors() {
			if _, err := FormatWithEntityFieldsSupport(first, supportedEntityFields()); err != nil {
				t.Fatalf("valid explicit parse cannot format: %v", err)
			}
		}
	})
}
