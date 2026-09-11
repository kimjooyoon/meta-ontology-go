package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"testing"
)

func TestSuffixTypeQualificationCohort(t *testing.T) {
	mainPackage := types.NewPackage("fixture/main", "main")
	localPackage := types.NewPackage("fixture/local", "sample")
	foreignPackage := types.NewPackage("fixture/foreign", "sample")
	mainType := suffixQualificationNamedType(mainPackage, "paginationFixtureFile")
	localType := suffixQualificationNamedType(localPackage, "Record")
	foreignType := suffixQualificationNamedType(foreignPackage, "Record")
	foreignPackage.MarkComplete()
	cases := []struct {
		name     string
		owner    *types.Package
		value    types.Type
		expected string
		imports  bool
	}{
		{"local-main", mainPackage, mainType, "paginationFixtureFile", false},
		{"foreign-same-name", localPackage, foreignType, "external.Record", true},
		{"mixed-local-and-foreign", localPackage, types.NewMap(localType, foreignType), "map[Record]external.Record", true},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			alias := ast.NewIdent("external")
			info := &types.Info{Uses: map[*ast.Ident]types.Object{
				alias: types.NewPkgName(token.NoPos, item.owner, alias.Name, foreignPackage),
			}}
			variable := types.NewVar(token.NoPos, item.owner, "value", item.value)
			fset := token.NewFileSet()
			bindings, err := renderSuffixBindings(map[types.Object]bool{variable: true}, fset, info, item.owner)
			if err != nil || len(bindings) != 1 {
				t.Fatalf("bindings=%v err=%v", bindings, err)
			}
			var rendered bytes.Buffer
			if err := format.Node(&rendered, fset, bindings[0].type_); err != nil {
				t.Fatal(err)
			}
			if rendered.String() != item.expected {
				t.Fatalf("type=%q want=%q", rendered.String(), item.expected)
			}
			checkSuffixQualificationType(t, item.owner.Name(), rendered.String(), foreignPackage, item.imports)
		})
	}
}

func suffixQualificationNamedType(pkg *types.Package, name string) *types.Named {
	object := types.NewTypeName(token.NoPos, pkg, name, nil)
	pkg.Scope().Insert(object)
	return types.NewNamed(object, types.NewStruct(nil, nil), nil)
}

func checkSuffixQualificationType(t *testing.T, packageName, expression string, foreign *types.Package, imports bool) {
	t.Helper()
	source := "package " + packageName + "\n"
	if imports {
		source += fmt.Sprintf("import external %q\n", foreign.Path())
	}
	source += "type Record struct{}\ntype paginationFixtureFile struct{}\nfunc accept(value " + expression + ") { _ = value }\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "qualification.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	configuration := types.Config{Importer: suffixQualificationImporter{foreign}}
	if _, err := configuration.Check("fixture/"+packageName, fset, []*ast.File{file}, nil); err != nil {
		t.Fatalf("rendered type does not type-check: %v", err)
	}
}

type suffixQualificationImporter struct {
	pkg *types.Package
}

func (importer suffixQualificationImporter) Import(path string) (*types.Package, error) {
	if path != importer.pkg.Path() {
		return nil, fmt.Errorf("unexpected fixture import %q", path)
	}
	return importer.pkg, nil
}
