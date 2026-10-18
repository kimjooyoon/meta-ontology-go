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
	if !ok || len(call.Args) != 2 || call.Ellipsis.IsValid() || f.Type.TypeParams != nil || f.Type.Params == nil || len(f.Type.Params.List) != 2 || f.Type.Results == nil || len(f.Type.Results.List) != 1 || len(f.Type.Results.List[0].Names) != 0 {
		return nil, nil, false
	}
	parameter := f.Type.Params.List[0]
	inputType, output := mapScalarSequence(parameter.Type), mapScalarSequence(f.Type.Results.List[0].Type)
	if len(parameter.Names) != 1 || parameter.Names[0].Name != "goooMapInput" || inputType == "" || output == "" || len(f.Body.List) != 4 {
		return nil, nil, false
	}
	apply := f.Type.Params.List[1]
	mapperType, ok := apply.Type.(*ast.FuncType)
	if !ok || len(apply.Names) != 1 || apply.Names[0].Name != "goooMapApply" || mapperType.TypeParams != nil || mapperType.Params == nil || len(mapperType.Params.List) != 1 || mapperType.Results == nil || len(mapperType.Results.List) != 1 || len(mapperType.Params.List[0].Names) != 0 || len(mapperType.Results.List[0].Names) != 0 || !mapName(mapperType.Params.List[0].Type, inputType) || !mapName(mapperType.Results.List[0].Type, output) {
		return nil, nil, false
	}
	mapper, ok := call.Args[1].(*ast.Ident)
	if !ok {
		return nil, nil, false
	}
	result, ok := mapDeclaration(f.Body.List[0], "goooMapResult")
	if !ok {
		return nil, nil, false
	}
	array, ok := result.(*ast.ArrayType)
	if !ok || !mapName(array.Elt, output) {
		return nil, nil, false
	}
	limit, ok := array.Len.(*ast.BasicLit)
	if !ok || limit.Kind != token.INT || limit.Value != "256" {
		return nil, nil, false
	}
	length, ok := mapDeclaration(f.Body.List[1], "goooMapLength")
	if !ok || !mapName(length, "int") {
		return nil, nil, false
	}
	loop, ok := f.Body.List[2].(*ast.RangeStmt)
	if !ok || loop.Tok != token.DEFINE || !mapName(loop.Key, "goooMapIndex") || loop.Value != nil || !mapName(loop.X, "goooMapInput") || len(loop.Body.List) != 2 {
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
	if !ok || !mapName(invoke.Fun, "goooMapApply") || len(invoke.Args) != 1 || invoke.Ellipsis.IsValid() {
		return nil, nil, false
	}
	read, ok := invoke.Args[0].(*ast.IndexExpr)
	if !ok || !mapName(read.X, "goooMapInput") || !mapName(read.Index, "goooMapIndex") {
		return nil, nil, false
	}
	update, ok := loop.Body.List[1].(*ast.AssignStmt)
	if !ok || update.Tok != token.ASSIGN || len(update.Lhs) != 1 || len(update.Rhs) != 1 || !mapName(update.Lhs[0], "goooMapLength") {
		return nil, nil, false
	}
	add, ok := update.Rhs[0].(*ast.BinaryExpr)
	if !ok || add.Op != token.ADD || !mapName(add.X, "goooMapIndex") {
		return nil, nil, false
	}
	one, ok := add.Y.(*ast.BasicLit)
	if !ok || one.Kind != token.INT || one.Value != "1" {
		return nil, nil, false
	}
	ret, ok := f.Body.List[3].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return nil, nil, false
	}
	slice, ok := ret.Results[0].(*ast.SliceExpr)
	if !ok || !slice.Slice3 || !mapName(slice.X, "goooMapResult") || slice.Low != nil || !mapName(slice.High, "goooMapLength") || !mapName(slice.Max, "goooMapLength") {
		return nil, nil, false
	}
	return call.Args[0], mapper, true
}

func mapDeclaration(statement ast.Stmt, name string) (ast.Expr, bool) {
	declaration, ok := statement.(*ast.DeclStmt)
	if !ok {
		return nil, false
	}
	general, ok := declaration.Decl.(*ast.GenDecl)
	if !ok || general.Tok != token.VAR || len(general.Specs) != 1 {
		return nil, false
	}
	variable, ok := general.Specs[0].(*ast.ValueSpec)
	if !ok || len(variable.Names) != 1 || variable.Names[0].Name != name || len(variable.Values) != 0 {
		return nil, false
	}
	return variable.Type, true
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
