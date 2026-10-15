package bodycodegen

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestLocalProjectionOrdinals10091(t *testing.T) {
	const body = "var subtotal = input + 83; subtotal = subtotal + 89; return subtotal"
	p, err := prepareBodyProjection("locals10091", "Adjust", []InputParameter{{Name: "input", Type: "int64"}}, "int64", body, nil)
	if err != nil {
		t.Fatal(err)
	}
	p.locals = p.captureLocalOrigins("urn:locals10091:adjust", body)
	// This is a compiler-created write, with no original source or Object.
	synthetic := &ast.DeclStmt{Decl: &ast.GenDecl{Tok: token.VAR, Specs: []ast.Spec{
		&ast.ValueSpec{Names: []*ast.Ident{ast.NewIdent("compilerTemporary")},
			Type: ast.NewIdent("int64"), Values: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "97"}}},
	}}}
	p.function.Body.List = append([]ast.Stmt{synthetic}, p.function.Body.List...)
	origins := p.emittedLocalOrigins()
	if len(origins) != 2 || origins[0].NativeOrdinal != 1 || origins[1].NativeOrdinal != 2 ||
		origins[0].VariableOrdinal != origins[1].VariableOrdinal {
		t.Fatal("synthetic write shifted authored variable identity", origins)
	}
	var source bytes.Buffer
	if err := format.Node(&source, p.fset, p.file); err != nil {
		t.Fatal(err)
	}
	reparsed, err := parser.ParseFile(token.NewFileSet(), "projection10091.go", source.Bytes(), 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	ast.Inspect(reparsed, func(node ast.Node) bool {
		if write := retainedLocalWrite(node); write.name != nil {
			names = append(names, write.name.Name)
		}
		return true
	})
	if len(names) != 3 || names[0] != "compilerTemporary" {
		t.Fatal("reparsed projection inventory", names)
	}
	for _, origin := range origins {
		if names[origin.NativeOrdinal] != origin.Name || origin.Name != "subtotal" {
			t.Fatal("native ordinal no longer names the authored variable", origin, names)
		}
	}
	saveLocalOrigin9811(t, "local10091.json", map[string]any{"normalized_body": body,
		"projection": source.String(), "origins": origins, "reparsed_names": names, "native_executions": 0})
}
