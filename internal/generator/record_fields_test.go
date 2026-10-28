package generator

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func recordFieldsFixture() SemanticIR {
	ir := entityFieldsFixture()
	ir.Activities = nil
	line := ir.Entities[0]
	line.ID, line.Name, line.GoName = "urn:projection:line", "Line", "NativeLine"
	for i := range line.Fields {
		line.Fields[i].ID, line.Fields[i].Parent = fmt.Sprintf("urn:projection:line:%d", i), line.ID
	}
	line.Fields[0].TypeRefID = entityFieldsIntegerTypeID
	basket := Entity{ID: "urn:projection:basket", Name: "Basket", GoName: "NativeBasket", Source: line.Source}
	for i, name := range []string{"One", "Optional", "Many", "OptionalMany"} {
		field := line.Fields[0]
		field.ID, field.Parent, field.Name, field.TypeRefID = fmt.Sprintf("urn:projection:basket:%d", i), basket.ID, name, line.ID
		field.Presence, field.Cardinality = []string{"required", "optional"}[i%2], []string{"one", "many"}[i/2]
		for _, span := range []*SourceSpan{&field.Source, &field.IDSpan, &field.NameSpan, &field.TypeRefSpan, &field.PresenceSpan, &field.CardinalitySpan} {
			span.Start.Offset += i * 20
			span.End.Offset += i * 20
			span.Start.Line += i
			span.End.Line += i
		}
		basket.Fields = append(basket.Fields, field)
	}
	// The native name is deliberately different from source presentation, and
	// the reference precedes its target declaration.
	ir.Entities = []Entity{basket, line}
	return ir
}

func TestRecordFieldsProjectAllShapesByStableTargetIdentity(t *testing.T) {
	ir := recordFieldsFixture()
	before := copyIR(ir)
	support, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	result, err := GenerateWithEntityFieldsSupport(ir, nil, support)
	if err != nil {
		t.Fatal(err)
	}
	metadata, err := generateProjectionV1WithEntityFieldsSupport(New(Options{}), ir, nil, support)
	if err != nil || metadata.SemanticIR.Entities[0].Fields[0].GoType != "NativeLine" ||
		metadata.Metadata.EntityFields == nil || metadata.Metadata.EntityFields.Profile.Digest != support.Profile.Digest {
		t.Fatal("versioned projection lost the nominal type/profile", err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "records.go", result.Source, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if ok && spec.Name.Name == "NativeBasket" {
			for _, field := range spec.Type.(*ast.StructType).Fields.List {
				names = append(names, field.Names[0].Name)
			}
		}
		return true
	})
	if strings.Join(names, ",") != "One,Optional,Many,OptionalMany" || !reflect.DeepEqual(ir, before) {
		t.Fatal("projection changed source field order or input IR", names)
	}
	for index, field := range ir.Entities[0].Fields {
		mappings := result.SourceMap.Lookup(field.ID)
		if len(mappings) != 1 || mappings[0].Kind != "field" || mappings[0].ParentID != field.Parent ||
			mappings[0].TypeRefID != "urn:projection:line" || mappings[0].Ordinal != index ||
			mappings[0].Source != field.Source || mappings[0].NameSource != field.NameSpan || mappings[0].ProfileDigest != support.Profile.Digest {
			t.Fatal("field projection lost stable target identity, source position or profile", mappings)
		}
	}
	again, err := GenerateWithEntityFieldsSupport(ir, result.Source, support)
	if err != nil || !bytes.Equal(again.Source, result.Source) {
		t.Fatal("unchanged projection is not idempotent", err)
	}
	renamed := copyIR(ir)
	renamed.Entities[1].Name, renamed.Entities[1].GoName = "Product", "NativeProduct"
	renameResult, err := GenerateWithEntityFieldsSupport(renamed, result.Source, support)
	if err != nil || bytes.Contains(renameResult.Source, []byte("NativeLine")) ||
		!bytes.Contains(renameResult.Source, []byte("*[]NativeProduct")) ||
		renameResult.SourceMap.Lookup(ir.Entities[0].Fields[0].ID)[0].TypeRefID != ir.Entities[1].ID {
		t.Fatal("target rename lost nominal identity or left a stale native type", err)
	}
	directory := t.TempDir()
	files := map[string][]byte{
		"go.mod": []byte("module recordprojection\n\ngo 1.27.3\n"), "records.go": result.Source,
		"records_test.go": []byte(`package entityfieldsgen
import "testing"
func TestNativeValues(t *testing.T) {
 line := NativeLine{OrderNumber: 9007199254740993, CustomerName: "가나다"}
 items := []NativeLine{line, line}
 basket := NativeBasket{One: line, Optional: &line, Many: items, OptionalMany: &items}
 if basket.One.OrderNumber != 9007199254740993 || basket.Optional.CustomerName != "가나다" ||
    len(basket.Many) != 2 || len(*basket.OptionalMany) != 2 { t.Fatal(basket) }
}


`),
	}
	for name, raw := range files {
		if err := os.WriteFile(filepath.Join(directory, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	command := exec.Command("go", "test", "./...")
	command.Dir = directory
	command.Env = append(os.Environ(), "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("independent native record value oracle: %v\n%s", err, output)
	}
}

func TestRecordFieldProjectionRefusesUnknownAndConflictingTypes(t *testing.T) {
	v7, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	for _, test := range []struct {
		name    string
		support syntax.EntityFieldsSupport
		edit    func(*SemanticIR)
		code    string
	}{
		{name: "older profile", support: syntax.EntityFieldsV6Support(), edit: func(*SemanticIR) {}, code: entityFieldsUnsupportedTypeDiagnostic},
		{name: "unknown ID", support: v7, edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].TypeRefID = "urn:missing" }, code: entityFieldsUnsupportedTypeDiagnostic},
		{name: "empty entity target", support: v7, edit: func(ir *SemanticIR) { ir.Entities[1].Fields = nil }, code: entityFieldsUnsupportedTypeDiagnostic},
		{name: "Go type override", support: v7, edit: func(ir *SemanticIR) { ir.Entities[0].Fields[0].GoType = "string" }, code: entityFieldsUnrepresentableDiagnostic},
		{name: "pointer cycle", support: v7, edit: func(ir *SemanticIR) {
			ir.Entities[0].Fields[0].Presence = "optional"
			ir.Entities[1].Fields[0].TypeRefID, ir.Entities[1].Fields[0].Presence = ir.Entities[0].ID, "optional"
		}, code: entityFieldsUnsupportedShapeDiagnostic},
		{name: "native scalar shadow", support: v7, edit: func(ir *SemanticIR) { ir.Entities[1].GoName = "string" }, code: entityFieldsGoNameCollisionDiagnostic},
		{name: "fieldless scalar shadow", support: v7, edit: func(ir *SemanticIR) {
			ir.Entities = append(ir.Entities, Entity{ID: "urn:projection:shadow", Name: "Shadow", GoName: "int64"})
		}, code: entityFieldsGoNameCollisionDiagnostic},
	} {
		t.Run(test.name, func(t *testing.T) {
			ir := recordFieldsFixture()
			test.edit(&ir)
			result, err := GenerateWithEntityFieldsSupport(ir, nil, test.support)
			if err == nil || !strings.Contains(err.Error(), test.code) || len(result.Source) != 0 {
				t.Fatal("unsupported nominal projection produced an output", err)
			}
		})
	}
}

func TestStandaloneRecordProjectionRetainsSourceDepthBoundary(t *testing.T) {
	support, _ := syntax.EntityFieldsSupportForProfileID("gooo.entityfields.go-projection.v7")
	for _, depth := range []int{8, 9} {
		ir := entityFieldsFixture()
		base := ir.Entities[0]
		ir.Entities, ir.Activities = nil, nil
		for i := range depth {
			entity := Entity{ID: fmt.Sprintf("urn:projection:depth:%d", i), Name: fmt.Sprintf("Layer%d", i), Source: base.Source}
			field := base.Fields[0]
			field.ID, field.Parent, field.Name = entity.ID+":next", entity.ID, "Next"
			field.TypeRefID = entityFieldsIntegerTypeID
			if i+1 < depth {
				field.TypeRefID = fmt.Sprintf("urn:projection:depth:%d", i+1)
			}
			entity.Fields = []Field{field}
			ir.Entities = append(ir.Entities, entity)
		}
		result, err := GenerateWithEntityFieldsSupport(ir, nil, support)
		if depth == 8 {
			if err != nil || len(result.Source) == 0 {
				t.Fatal("source-admitted nominal depth was rejected", err)
			}
		} else if err == nil || !strings.Contains(err.Error(), "record depth exceeds 8") ||
			!strings.Contains(err.Error(), "urn:projection:depth:7:next") || len(result.Source) != 0 {
			t.Fatal("standalone generator bypassed the shared source depth/field contract", err)
		}
	}
}
