package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
)

func candidates(fset *token.FileSet, file *ast.File) ([]declaration, bool, error) {
	seen := map[string]bool{}
	var out []declaration
	fallbackUsed := false
	for order, node := range file.Decls {
		identity, movable := identityOf(fset, node)
		if !movable { identity, movable = fallbackIdentity(node); fallbackUsed = fallbackUsed || movable }
		if !movable { continue }
		if seen[identity] {
			return nil, false, fail("validate-ast", "identity", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{identity})
		}
		seen[identity] = true
		start := node.Pos()
		if d, ok := node.(*ast.FuncDecl); ok && d.Doc != nil { start = d.Doc.Pos() }
		if d, ok := node.(*ast.GenDecl); ok && d.Doc != nil { start = d.Doc.Pos() }
		out = append(out, declaration{node, fset.Position(start).Offset, fset.Position(node.End()).Offset, order, identity})
	}
	return out, fallbackUsed, nil
}

func fallbackIdentity(node ast.Decl) (string, bool) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" { return "", false }
		return fmt.Sprintf("func-at:%d", d.Pos()), true
	case *ast.GenDecl:
		if d.Tok == token.IMPORT { return "", false }
		return fmt.Sprintf("group-at:%d", d.Pos()), true
	default:
		return "", false
	}
}

func identityOf(fset *token.FileSet, node ast.Decl) (string, bool) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" { return "", false }
		if d.Recv == nil { return "func:" + d.Name.Name, true }
		var receiver bytes.Buffer
		if format.Node(&receiver, fset, d.Recv) != nil { return "", false }
		return "method:" + receiver.String() + ":" + d.Name.Name, true
	case *ast.GenDecl:
		if d.Tok == token.IMPORT { return "", false }
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
