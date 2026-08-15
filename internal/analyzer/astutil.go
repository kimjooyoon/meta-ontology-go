package analyzer

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
)

type importTable struct {
	aliases map[string]string
	dot     []string
}

func importsFor(file *ast.File) importTable {
	table := importTable{aliases: make(map[string]string)}
	for _, specification := range file.Imports {
		path, err := strconv.Unquote(specification.Path.Value)
		if err != nil || path == "" {
			continue
		}
		if specification.Name != nil {
			switch specification.Name.Name {
			case "_":
				continue
			case ".":
				table.dot = append(table.dot, path)
			default:
				table.aliases[specification.Name.Name] = path
			}
			continue
		}
		alias := path
		if slash := strings.LastIndex(alias, "/"); slash >= 0 {
			alias = alias[slash+1:]
		}
		table.aliases[alias] = path
	}
	return table
}

type typeReference struct {
	packagePath string
	packageName string
	name        string
}

func variableTypes(function *ast.FuncDecl, file parsedFile, imports importTable) map[string]typeReference {
	result := make(map[string]typeReference)
	collectVariableTypes(result, function.Recv, file, imports)
	collectVariableTypes(result, function.Type.Params, file, imports)
	if function.Body != nil {
		ast.Inspect(function.Body, func(node ast.Node) bool {
			declaration, ok := node.(*ast.ValueSpec)
			if ok && declaration.Type != nil {
				collectVariableType(result, declaration, file, imports)
			}
			return true
		})
	}
	return result
}

func collectVariableTypes(result map[string]typeReference, fields *ast.FieldList, file parsedFile, imports importTable) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		ref, ok := typeReferenceFor(field.Type, file, imports)
		if ok {
			for _, name := range field.Names {
				result[name.Name] = ref
			}
		}
	}
}

func collectVariableType(result map[string]typeReference, declaration *ast.ValueSpec, file parsedFile, imports importTable) {
	ref, ok := typeReferenceFor(declaration.Type, file, imports)
	if !ok {
		return
	}
	for _, name := range declaration.Names {
		result[name.Name] = ref
	}
}

func typeReferenceFor(expr ast.Expr, file parsedFile, imports importTable) (typeReference, bool) {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return typeReference{packagePath: file.packagePath, packageName: file.packageName, name: current.Name}, true
	case *ast.SelectorExpr:
		base, ok := unwrapExpr(current.X).(*ast.Ident)
		if !ok {
			return typeReference{}, false
		}
		path, ok := imports.aliases[base.Name]
		if !ok {
			return typeReference{}, false
		}
		return typeReference{packagePath: path, packageName: base.Name, name: current.Sel.Name}, true
	case *ast.StarExpr:
		return typeReferenceFor(current.X, file, imports)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return typeReferenceFor(indexBase(expr), file, imports)
	default:
		return typeReference{}, false
	}
}

func receiverTypeName(fields *ast.FieldList) string {
	if fields == nil || len(fields.List) == 0 {
		return ""
	}
	return typeName(fields.List[0].Type)
}

func typeName(expr ast.Expr) string {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return current.Name
	case *ast.StarExpr:
		return typeName(current.X)
	case *ast.IndexExpr, *ast.IndexListExpr:
		return typeName(indexBase(expr))
	default:
		return ""
	}
}

func unwrapExpr(expr ast.Expr) ast.Expr {
	for expr != nil {
		switch current := expr.(type) {
		case *ast.ParenExpr:
			expr = current.X
		case *ast.IndexExpr, *ast.IndexListExpr:
			expr = indexBase(expr)
		default:
			return expr
		}
	}
	return nil
}

func indexBase(expr ast.Expr) ast.Expr {
	switch current := expr.(type) {
	case *ast.IndexExpr:
		return current.X
	case *ast.IndexListExpr:
		return current.X
	default:
		return expr
	}
}

func expressionName(expr ast.Expr) string {
	switch current := unwrapExpr(expr).(type) {
	case *ast.Ident:
		return current.Name
	case *ast.SelectorExpr:
		left := expressionName(current.X)
		if left == "" {
			return current.Sel.Name
		}
		return left + "." + current.Sel.Name
	default:
		return "<expression>"
	}
}

func spanFor(fileSet *token.FileSet, node ast.Node) Span {
	if node == nil {
		return Span{}
	}
	start := fileSet.PositionFor(node.Pos(), true)
	end := fileSet.PositionFor(node.End(), true)
	return Span{
		Filename: start.Filename,
		Start:    Position{Offset: start.Offset, Line: start.Line, Column: start.Column},
		End:      Position{Offset: end.Offset, Line: end.Line, Column: end.Column},
	}
}
