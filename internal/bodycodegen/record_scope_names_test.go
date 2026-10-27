package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

func TestRecordScopeSeparatesSourceAndNativeNamingViews(t *testing.T) {
	for _, count := range []int{16, 17} {
		model := bidir.Model{}
		for index := range count {
			model.Nodes = append(model.Nodes, recordShapeFixture(fmt.Sprintf("R%d", index), 1))
		}
		prefix, err := DeclaredRecordTypesFromModel(model)
		if err != nil {
			t.Fatal(err)
		}
		// A source display name can equal another record's generated spelling.
		model.Nodes = append(model.Nodes, recordShapeFixture(prefix[0].GoName, 1))
		records, err := DeclaredRecordTypesFromModel(model)
		if err != nil {
			t.Fatal(err)
		}
		for _, native := range []bool{false, true} {
			source := "package names\n" + RecordDeclarations(records, native) + "func Body(){\n"
			for _, record := range records[:count] {
				name, field := record.Name, record.Fields[0].Name
				if native {
					name, field = record.GoName, record.Fields[0].GoName
				}
				source += "_ = " + name + "{" + field + ":0}\n"
			}
			source += "}\n"
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "names.go", source, 0)
			if err != nil {
				t.Fatal(err)
			}
			info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
			if _, err := new(types.Config).Check("names", fset, []*ast.File{file}, info); err != nil {
				t.Fatal(err)
			}
			err = ValidateRecordUseBudget(info, records, native)
			if count == 16 && err != nil || count == 17 && (err == nil || !strings.Contains(err.Error(), "supports at most 16")) {
				t.Fatal("naming view changed the source-owned count", count, native, err)
			}
		}
	}
}
