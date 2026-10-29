package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/token"
)

func checkAdmissionRootBindings(source *goSource, decode, header *ast.FuncDecl) error {
	if !admissionRootParameter(source, decode, "raw", "[]byte") ||
		!admissionRootParameter(source, header, "value", "denominator") ||
		len(decode.Body.List) == 0 || source.text(decode.Body.List[0]) != "digest := digestBytes(raw)" {
		return fmt.Errorf("denominator admission input binding differs")
	}
	for _, function := range []*ast.FuncDecl{decode, header} {
		var names []*ast.Ident
		if function.Type.Results != nil {
			for _, field := range function.Type.Results.List {
				names = append(names, field.Names...)
			}
		}
		for _, statement := range function.Body.List {
			names = append(names, admissionStatementNames(statement)...)
		}
		for _, name := range names {
			if _, reserved := digestVersion(name.Name); reserved {
				return fmt.Errorf("denominator admission digest binding is locally shadowed: %s", name.Name)
			}
		}
	}
	return nil
}

func admissionRootParameter(source *goSource, function *ast.FuncDecl, name, kind string) bool {
	if function.Type.TypeParams != nil || function.Type.Params.NumFields() != 1 || function.Body == nil {
		return false
	}
	parameter := function.Type.Params.List[0]
	return len(parameter.Names) == 1 && parameter.Names[0].Name == name && source.text(parameter.Type) == kind
}

func admissionStatementNames(statement ast.Stmt) []*ast.Ident {
	var names []*ast.Ident
	if assignment, ok := statement.(*ast.AssignStmt); ok && assignment.Tok == token.DEFINE {
		for _, expression := range assignment.Lhs {
			if name, named := expression.(*ast.Ident); named {
				names = append(names, name)
			}
		}
	}
	if declaration, ok := statement.(*ast.DeclStmt); ok {
		if group, grouped := declaration.Decl.(*ast.GenDecl); grouped {
			for _, spec := range group.Specs {
				switch spec := spec.(type) {
				case *ast.ValueSpec:
					names = append(names, spec.Names...)
				case *ast.TypeSpec:
					names = append(names, spec.Name)
				}
			}
		}
	}
	return names
}
