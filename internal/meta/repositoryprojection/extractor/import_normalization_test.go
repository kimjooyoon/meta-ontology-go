package extractor

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

func TestImportNormalizationFixedRegressionCohort(t *testing.T) {
	cases := []struct {
		name        string
		source      string
		normalized  bool
		wantImports string
	}{
		{name: "eligible-two-spec-group-normalizes", source: "package p\n\nimport (\n\t\"encoding/json\"\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n", normalized: true, wantImports: "import \"encoding/json\"\nimport \"strconv\"\n"},
		{name: "one-spec-import-stays-unchanged", source: "package p\n\nimport \"encoding/json\"\n\nfunc F() {\n\tvar _ json.RawMessage\n}\n", normalized: false, wantImports: "import \"encoding/json\"\n"},
	}
	if len(cases) != 2 {
		t.Fatalf("import normalization cohort denominator=%d, want 2", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			helper, err := renderImportFixture(t, tc.source, "F")
			if err != nil || string(helper) != tc.wantImports {
				t.Fatalf("helper imports=%q err=%v", helper, err)
			}
			if tc.normalized && strings.Contains(string(helper), "import (\n") {
				t.Fatalf("eligible import group retained grouped rendering: %q", helper)
			}
		})
	}
}

func TestImportNormalizationPreservationRegressionCohort(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{name: "alias", source: "package p\n\nimport (\n\tj \"encoding/json\"\n\t\"strconv\"\n)\n"},
		{name: "dot", source: "package p\n\nimport (\n\t. \"encoding/json\"\n\t\"strconv\"\n)\n"},
		{name: "blank", source: "package p\n\nimport (\n\t_ \"encoding/json\"\n\t\"strconv\"\n)\n"},
		{name: "group-comment", source: "package p\n\n// retain import documentation\nimport (\n\t\"encoding/json\"\n\t\"strconv\"\n)\n"},
		{name: "spec-comment", source: "package p\n\nimport (\n\t\"encoding/json\" // retain import comment\n\t\"strconv\"\n)\n"},
		{name: "cgo", source: "package p\n\nimport (\n\t\"C\"\n\t\"strconv\"\n)\n"},
	}
	if len(cases) != 6 {
		t.Fatalf("import preservation cohort denominator=%d, want 6", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			file := importFixtureFile(t, tc.source)
			group := firstImportGroupInFile(t, file)
			if eligiblePlainImportGroup(file, group, importSpecsForGroup(group)) {
				t.Fatal("unsupported import group was admitted to normalization")
			}
		})
	}
}

func TestImportNormalizationCorrectionRegressionCohort(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{name: "detached-comment-preserves-grouped-rendering", source: "package p\n\nimport (\n\t\"encoding/json\"\n\n\t// detached import note\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n"},
		{name: "detached-directive-preserves-grouped-rendering", source: "package p\n\nimport (\n\t\"encoding/json\"\n\n\t//go:custom-import-directive\n\t\"strconv\"\n)\n\nfunc F() {\n\tvar _ json.RawMessage\n\t_ = strconv.IntSize\n}\n"},
		{name: "raw-string-c-preserves-grouped-rendering", source: "package p\n\nimport (\n\t`C`\n\t\"strconv\"\n)\n\nfunc F() {\n\t_ = C.symbol\n\t_ = strconv.IntSize\n}\n"},
	}
	if len(cases) != 3 {
		t.Fatalf("import normalization correction cohort denominator=%d, want 3", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			helper, err := renderImportFixtureWithRawImportList(t, tc.source, "F")
			if err != nil || !strings.Contains(string(helper), "import (\n") {
				t.Fatalf("preservation renderer=%q err=%v", helper, err)
			}
		})
	}
}

func TestImportNormalizationDeterminismAndMethodIdentity(t *testing.T) {
	source := "package p\n\nimport (\n\t\"encoding/json\"\n\t\"strconv\"\n)\n\ntype T struct{}\n\nfunc (T) F(v json.RawMessage) error {\n\t_ = strconv.IntSize\n\treturn nil\n}\n"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	list, err := imports(file)
	if err != nil {
		t.Fatal(err)
	}
	var method *ast.FuncDecl
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv != nil && function.Name.Name == "F" {
			method = function
		}
	}
	if method == nil {
		t.Fatal("method fixture missing")
	}
	start, end := fset.Position(method.Pos()).Offset, fset.Position(method.End()).Offset
	selected := []declaration{{node: method, start: start, end: end, identity: "method:T:F"}}
	first, err := render(fset, file, []byte(source), selected, list)
	if err != nil {
		t.Fatal(err)
	}
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	list, err = imports(file)
	if err != nil {
		t.Fatal(err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Recv != nil && function.Name.Name == "F" {
			method = function
		}
	}
	start, end = fset.Position(method.Pos()).Offset, fset.Position(method.End()).Offset
	selected = []declaration{{node: method, start: start, end: end, identity: "method:T:F"}}
	second, err := render(fset, file, []byte(source), selected, list)
	if err != nil || !bytes.Equal(first.helper, second.helper) {
		t.Fatalf("normalization was not deterministic: first=%q second=%q err=%v", first.helper, second.helper, err)
	}
	if !strings.Contains(string(first.helper), "func (T) F(") || !strings.Contains(string(first.helper), "_ = strconv.IntSize") {
		t.Fatalf("method body or receiver identity changed: %q", first.helper)
	}
}

func renderImportFixture(t *testing.T, source, name string) ([]byte, error) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		return nil, err
	}
	list, err := imports(file)
	if err != nil {
		return nil, err
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == name {
			_, helper, renderErr := renderImports(fset, file, []ast.Decl{function}, list, false)
			return helper, renderErr
		}
	}
	return nil, nil
}

func renderImportFixtureWithRawImportList(t *testing.T, source, name string) ([]byte, error) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		return nil, err
	}
	group := firstImportGroupInFile(t, file)
	list := make([]importSpec, 0, len(group.Specs))
	for _, raw := range group.Specs {
		spec := raw.(*ast.ImportSpec)
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			path = spec.Path.Value
		}
		importedName := ""
		if spec.Name != nil {
			importedName = spec.Name.Name
		}
		list = append(list, importSpec{group: group, spec: spec, path: path, name: importedName})
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && function.Name != nil && function.Name.Name == name {
			_, helper, renderErr := renderImports(fset, file, []ast.Decl{function}, list, false)
			return helper, renderErr
		}
	}
	return nil, nil
}

func importFixtureFile(t *testing.T, source string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "fixture.go", []byte(source), parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func firstImportGroupInFile(t *testing.T, file *ast.File) *ast.GenDecl {
	t.Helper()
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if ok && group.Tok == token.IMPORT {
			return group
		}
	}
	t.Fatal("import group missing")
	return nil
}

func importSpecsForGroup(group *ast.GenDecl) []*ast.ImportSpec {
	result := make([]*ast.ImportSpec, 0, len(group.Specs))
	for _, raw := range group.Specs {
		if spec, ok := raw.(*ast.ImportSpec); ok {
			result = append(result, spec)
		}
	}
	return result
}
