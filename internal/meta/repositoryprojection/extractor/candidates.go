package extractor

import (
	"fmt"
	"go/ast"
	"go/token"
)

func candidates(fset *token.FileSet, file *ast.File) ([]declaration, bool, error) {
	seen := map[string]bool{}
	var out []declaration
	fallbackUsed := false
	for order, node := range file.Decls {
		identity, movable := identityOf(fset, node)
		if !movable {
			if function, ok := node.(*ast.FuncDecl); ok && function.Recv != nil && function.Name != nil && function.Name.Name != "init" && function.Name.Name != "_" {
				return nil, false, failWithDiagnostics("validate-ast", "identity", "UNSUPPORTED_RECEIVER", "KNOWN_CONTRADICTION", "report-contradiction", []string{"method=" + function.Name.Name})
			}
			identity, movable = fallbackIdentity(node)
			fallbackUsed = fallbackUsed || movable
		}
		if !movable {
			continue
		}
		if seen[identity] {
			return nil, false, fail("validate-ast", "identity", "DECLARATION_IDENTITY_COLLISION", "KNOWN_CONTRADICTION", "report-contradiction", []string{identity})
		}
		seen[identity] = true
		start := node.Pos()
		if d, ok := node.(*ast.FuncDecl); ok && d.Doc != nil {
			start = d.Doc.Pos()
		}
		if d, ok := node.(*ast.GenDecl); ok && d.Doc != nil {
			start = d.Doc.Pos()
		}
		out = append(out, declaration{node, fset.Position(start).Offset, fset.Position(node.End()).Offset, order, identity})
	}
	return out, fallbackUsed, nil
}

func fallbackIdentity(node ast.Decl) (string, bool) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" {
			return "", false
		}
		return fmt.Sprintf("func-at:%d", d.Pos()), true
	case *ast.GenDecl:
		if d.Tok == token.IMPORT {
			return "", false
		}
		return fmt.Sprintf("group-at:%d", d.Pos()), true
	default:
		return "", false
	}
}

func identityOf(fset *token.FileSet, node ast.Decl) (string, bool) {
	switch d := node.(type) {
	case *ast.FuncDecl:
		if d.Name == nil || d.Name.Name == "init" || d.Name.Name == "_" {
			return "", false
		}
		if d.Recv == nil {
			return "func:" + d.Name.Name, true
		}
		receiver, ok := receiverBaseIdentifier(d.Recv)
		if !ok {
			return "", false
		}
		return "method:" + receiver + ":" + d.Name.Name, true
	case *ast.GenDecl:
		if d.Tok == token.IMPORT {
			return "", false
		}
		for _, raw := range d.Specs {
			name := ""
			switch spec := raw.(type) {
			case *ast.TypeSpec:
				name = spec.Name.Name
			case *ast.ValueSpec:
				if len(spec.Names) > 0 {
					name = spec.Names[0].Name
				}
			}
			if name != "" {
				return d.Tok.String() + ":" + name, true
			}
		}
	}
	return "", false
}

// receiverBaseIdentifier returns the syntactic base type name used by a
// method declaration. It is a declaration coordinate, not semantic alias
// resolution: pointer/value form and receiver type-parameter binder names do
// not change the key.
func receiverBaseIdentifier(receiver *ast.FieldList) (string, bool) {
	if receiver == nil || len(receiver.List) != 1 {
		return "", false
	}
	field := receiver.List[0]
	if field == nil || field.Type == nil || len(field.Names) > 1 {
		return "", false
	}
	return receiverBaseTypeIdentifier(field.Type)
}

func receiverBaseTypeIdentifier(expression ast.Expr) (string, bool) {
	if parenthesized, ok := expression.(*ast.ParenExpr); ok {
		if parenthesized == nil {
			return "", false
		}
		return receiverBaseTypeIdentifier(parenthesized.X)
	}
	if pointer, ok := expression.(*ast.StarExpr); ok {
		if pointer == nil {
			return "", false
		}
		return receiverBaseTypeIdentifierWithoutPointer(pointer.X)
	}
	return receiverBaseTypeIdentifierWithoutPointer(expression)
}

func receiverBaseTypeIdentifierWithoutPointer(expression ast.Expr) (string, bool) {
	switch value := expression.(type) {
	case *ast.Ident:
		if value == nil || value.Name == "" || value.Name == "_" {
			return "", false
		}
		return value.Name, true
	case *ast.ParenExpr:
		if value == nil {
			return "", false
		}
		return receiverBaseTypeIdentifierWithoutPointer(value.X)
	case *ast.IndexExpr:
		if value == nil {
			return "", false
		}
		return indexedReceiverBaseIdentifier(value.X, []ast.Expr{value.Index})
	case *ast.IndexListExpr:
		if value == nil {
			return "", false
		}
		return indexedReceiverBaseIdentifier(value.X, value.Indices)
	default:
		return "", false
	}
}

func indexedReceiverBaseIdentifier(base ast.Expr, arguments []ast.Expr) (string, bool) {
	identifier, ok := base.(*ast.Ident)
	if !ok || identifier == nil || identifier.Name == "" || identifier.Name == "_" || len(arguments) == 0 {
		return "", false
	}
	for _, argument := range arguments {
		binder, ok := argument.(*ast.Ident)
		if !ok || binder == nil || binder.Name == "" {
			return "", false
		}
	}
	return identifier.Name, true
}
