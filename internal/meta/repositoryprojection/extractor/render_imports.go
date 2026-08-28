package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

func renderImports(fset *token.FileSet, file *ast.File, decls []ast.Decl, list []importSpec, includeBlank bool) (map[*ast.GenDecl][]byte, []byte, error) {
	selected, err := selectedImports(decls, list, includeBlank); if err != nil { return nil, nil, err }
	replacements := map[*ast.GenDecl][]byte{}; var helper bytes.Buffer
	for _, node := range file.Decls {
		group, ok := node.(*ast.GenDecl); if !ok || group.Tok != token.IMPORT || len(selected[group]) == 0 { continue }
		data, err := formatImport(fset, group, selected[group]); if err != nil { return nil, nil, err }
		replacements[group] = data; helper.Write(data); helper.WriteByte('\n')
	}
	return replacements, helper.Bytes(), nil
}

func formatImport(fset *token.FileSet, group *ast.GenDecl, specs []*ast.ImportSpec) ([]byte, error) {
	copyGroup := *group; copyGroup.Specs = make([]ast.Spec, len(specs)); for i, spec := range specs { copyGroup.Specs[i] = spec }
	if len(specs) == 1 { copyGroup.Lparen, copyGroup.Rparen = token.NoPos, token.NoPos }
	var out bytes.Buffer; if err := format.Node(&out, fset, &copyGroup); err != nil { return nil, fail("rewrite-source", "render-imports", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil) }
	return out.Bytes(), nil
}
