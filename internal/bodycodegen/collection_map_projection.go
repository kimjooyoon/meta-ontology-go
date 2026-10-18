package bodycodegen

import (
	"go/ast"
	"go/token"
)

func mapName(expression ast.Expr, name string) bool {
	value, ok := expression.(*ast.Ident)
	return ok && value.Name == name
}

func mapScalarSequence(expression ast.Expr) string {
	sequence, ok := expression.(*ast.ArrayType)
	if !ok || sequence.Len != nil {
		return ""
	}
	name, ok := sequence.Elt.(*ast.Ident)
	if !ok || (name.Name != "string" && name.Name != "int64" && name.Name != "bool") {
		return ""
	}
	return name.Name
}

// CollectionMapProjection recognizes the complete closed lowering shape. The
// inverse rule checks allocation, traversal order, both indices and the return;
// recognizing only a generated name would permit a changed mapper or traversal.
func CollectionMapProjection(call *ast.CallExpr) (ast.Expr, *ast.Ident, bool) {
	f, ok := call.Fun.(*ast.FuncLit)
	if !ok || len(call.Args) != 1 || call.Ellipsis.IsValid() || f.Type.TypeParams != nil ||
		f.Type.Params == nil || len(f.Type.Params.List) != 1 || f.Type.Results == nil || len(f.Type.Results.List) != 1 {
		return nil, nil, false
	}
	parameter := f.Type.Params.List[0]
	output := mapScalarSequence(f.Type.Results.List[0].Type)
	if len(parameter.Names) != 1 || parameter.Names[0].Name != "goooMapInput" ||
		mapScalarSequence(parameter.Type) == "" || output == "" || len(f.Body.List) != 3 {
		return nil, nil, false
	}
	declaration, ok := f.Body.List[0].(*ast.DeclStmt)
	if !ok {
		return nil, nil, false
	}
	general, ok := declaration.Decl.(*ast.GenDecl)
	if !ok || general.Tok != token.VAR || len(general.Specs) != 1 {
		return nil, nil, false
	}
	variable, ok := general.Specs[0].(*ast.ValueSpec)
	if !ok || len(variable.Names) != 1 || variable.Names[0].Name != "goooMapResult" ||
		variable.Type != nil || len(variable.Values) != 1 {
		return nil, nil, false
	}
	allocation, ok := variable.Values[0].(*ast.CallExpr)
	if !ok || !mapName(allocation.Fun, "make") || len(allocation.Args) != 2 || allocation.Ellipsis.IsValid() ||
		mapScalarSequence(allocation.Args[0]) != output {
		return nil, nil, false
	}
	length, ok := allocation.Args[1].(*ast.CallExpr)
	if !ok || !mapName(length.Fun, "len") || len(length.Args) != 1 ||
		!mapName(length.Args[0], "goooMapInput") || length.Ellipsis.IsValid() {
		return nil, nil, false
	}
	loop, ok := f.Body.List[1].(*ast.RangeStmt)
	if !ok || loop.Tok != token.DEFINE || !mapName(loop.Key, "goooMapIndex") || loop.Value != nil ||
		!mapName(loop.X, "goooMapInput") || len(loop.Body.List) != 1 {
		return nil, nil, false
	}
	write, ok := loop.Body.List[0].(*ast.AssignStmt)
	if !ok || write.Tok != token.ASSIGN || len(write.Lhs) != 1 || len(write.Rhs) != 1 {
		return nil, nil, false
	}
	target, ok := write.Lhs[0].(*ast.IndexExpr)
	if !ok || !mapName(target.X, "goooMapResult") || !mapName(target.Index, "goooMapIndex") {
		return nil, nil, false
	}
	invoke, ok := write.Rhs[0].(*ast.CallExpr)
	if !ok || len(invoke.Args) != 1 || invoke.Ellipsis.IsValid() {
		return nil, nil, false
	}
	mapper, ok := invoke.Fun.(*ast.Ident)
	if !ok {
		return nil, nil, false
	}
	read, ok := invoke.Args[0].(*ast.IndexExpr)
	if !ok || !mapName(read.X, "goooMapInput") || !mapName(read.Index, "goooMapIndex") {
		return nil, nil, false
	}
	result, ok := f.Body.List[2].(*ast.ReturnStmt)
	if !ok || len(result.Results) != 1 || !mapName(result.Results[0], "goooMapResult") {
		return nil, nil, false
	}
	return call.Args[0], mapper, true
}

func restoreCollectionMaps(file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if input, mapper, recognized := CollectionMapProjection(call); recognized {
				call.Fun = ast.NewIdent("map_values")
				call.Args = []ast.Expr{input, ast.NewIdent(mapper.Name)}
			}
		}
		return true
	})
}

func hasCollectionMaps(file *ast.File) bool {
	if _, shadowed := findFunction(file, "map_values"); shadowed {
		return false
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && mapName(call.Fun, "map_values") {
			found = true
		}
		return !found
	})
	return found
}
