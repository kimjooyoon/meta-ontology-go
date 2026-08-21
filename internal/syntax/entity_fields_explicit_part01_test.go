package syntax

import (
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
