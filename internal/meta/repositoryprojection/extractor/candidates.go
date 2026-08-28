package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/token"
)

func candidates(fset *token.FileSet, file *ast.File) ([]declaration, error) {
	seen := map[string]bool{}
	var out []declaration
	for order, node := range file.Decls {
		identity, movable := identityOf(fset, node)
		if !movable { continue }
		if seen[identity] {
			return nil, fail("validate-ast", "identity", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{identity})
		}
		seen[identity] = true
		start := node.Pos()
		if d, ok := node.(*ast.FuncDecl); ok && d.Doc != nil { start = d.Doc.Pos() }
		if d, ok := node.(*ast.GenDecl); ok && d.Doc != nil { start = d.Doc.Pos() }
		out = append(out, declaration{node, fset.Position(start).Offset, fset.Position(node.End()).Offset, order, identity})
	}
	return out, nil
}

func identityOf(fset *token.FileSet, node ast.Decl) (string, bool) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" || d.Body == nil { return "", false }
		if d.Recv == nil { return "func:" + d.Name.Name, true }
		var receiver bytes.Buffer
		if format.Node(&receiver, fset, d.Recv) != nil { return "", false }
		return "method:" + receiver.String() + ":" + d.Name.Name, true
	case *ast.GenDecl:
		if d.Tok != token.CONST && d.Tok != token.TYPE { return "", false }
		for _, raw := range d.Specs {
			name := ""
			switch spec := raw.(type) {
			case *ast.TypeSpec: name = spec.Name.Name
			case *ast.ValueSpec: if len(spec.Names) > 0 { name = spec.Names[0].Name }
			}
			if name != "" { return d.Tok.String() + ":" + name, true }
		}
	}
	return "", false
}
