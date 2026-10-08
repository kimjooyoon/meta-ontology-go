package bodycodegen

import (
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"slices"
)

// Gooo locals keep their initialization even when no selected path reads them.
// Temporary reads let Go check all types without imposing its unused-local rule
// on the source view. The original nodes and positions are restored afterwards.
func checkBodyTypes(packageName string, file *ast.File, fset *token.FileSet, info *types.Info) (*types.Package, error) {
	restore := insertLocalReads(file, func(*ast.Ident) bool { return true })
	defer restore()
	configuration := types.Config{Importer: importer.Default()}
	return configuration.Check(packageName, fset, []*ast.File{file}, info)
}

func insertLocalReads(file *ast.File, include func(*ast.Ident) bool) func() {
	originals := make(map[*ast.BlockStmt][]ast.Stmt)
	ast.Inspect(file, func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		var statements []ast.Stmt
		for _, statement := range block.List {
			statements = append(statements, statement)
			for _, name := range declaredLocalNames(statement) {
				if name.Name != "_" && include(name) {
					statements = append(statements, &ast.AssignStmt{Tok: token.ASSIGN,
						Lhs: []ast.Expr{ast.NewIdent("_")}, Rhs: []ast.Expr{ast.NewIdent(name.Name)}})
				}
			}
		}
		if len(statements) != len(block.List) {
			originals[block], block.List = block.List, statements
		}
		return true
	})
	return func() {
		for block, statements := range originals {
			block.List = statements
		}
	}
}

func declaredLocalNames(statement ast.Stmt) []*ast.Ident {
	declaration, ok := statement.(*ast.DeclStmt)
	if !ok {
		return nil
	}
	general, ok := declaration.Decl.(*ast.GenDecl)
	if !ok || general.Tok != token.VAR || len(general.Specs) != 1 {
		return nil
	}
	value, ok := general.Specs[0].(*ast.ValueSpec)
	if !ok {
		return nil
	}
	return value.Names
}

// Only the native projection retains these reads. Initializers and subsequent
// assignments stay in their original scope and order, including possible errors.
func emitUnusedLocalReads(packageName string, file *ast.File, fset *token.FileSet) error {
	info := &types.Info{Defs: make(map[*ast.Ident]types.Object), Uses: make(map[*ast.Ident]types.Object)}
	if _, err := checkBodyTypes(packageName, file, fset, info); err != nil {
		return err
	}
	writes := make(map[*ast.Ident]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if assignment, ok := node.(*ast.AssignStmt); ok {
			for _, target := range assignment.Lhs {
				if name, ok := target.(*ast.Ident); ok {
					writes[name] = true
				}
			}
		}
		return true
	})
	reads := make(map[types.Object]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if name, ok := node.(*ast.Ident); ok && !writes[name] && info.Uses[name] != nil {
			reads[info.Uses[name]] = true
		}
		return true
	})
	insertLocalReads(file, func(name *ast.Ident) bool { return info.Defs[name] != nil && !reads[info.Defs[name]] })
	configuration := types.Config{Importer: importer.Default()}
	_, err := configuration.Check(packageName, fset, []*ast.File{file}, nil)
	return err
}

// A read immediately after its own declaration is the target's no-op usage
// marker. Removing only that shape keeps initializers and real assignments in
// source/target equivalence; arbitrary discarded expressions are never removed.
func removeLocalReadMarkers(file *ast.File, fset *token.FileSet) {
	lines := make(map[*token.File]map[int]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		block, ok := node.(*ast.BlockStmt)
		if !ok {
			return true
		}
		var statements []ast.Stmt
		for i := 0; i < len(block.List); i++ {
			statement := block.List[i]
			statements = append(statements, statement)
			names := declaredLocalNames(statement)
			if len(names) == 1 && i+1 < len(block.List) && isLocalReadMarker(block.List[i+1], names[0].Name) {
				marker := block.List[i+1]
				if source := fset.File(marker.Pos()); source != nil {
					line := source.Line(marker.Pos())
					if line > source.Line(statement.End()) && line < source.LineCount() {
						if lines[source] == nil {
							lines[source] = make(map[int]bool)
						}
						lines[source][line] = true
					}
				}
				i++
			}
		}
		block.List = statements
		return true
	})
	mergeRemovedMarkerLines(lines)
}

// Removing a generated read must not leave a synthetic blank line in the
// formatted semantic form. Source byte positions and literal text stay intact.
func mergeRemovedMarkerLines(lines map[*token.File]map[int]bool) {
	for source, removed := range lines {
		ordered := make([]int, 0, len(removed))
		for line := range removed {
			ordered = append(ordered, line)
		}
		slices.Sort(ordered)
		for _, line := range slices.Backward(ordered) {
			source.MergeLine(line)
		}
	}
}

func isLocalReadMarker(statement ast.Stmt, local string) bool {
	assignment, ok := statement.(*ast.AssignStmt)
	if !ok || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return false
	}
	left, leftOK := assignment.Lhs[0].(*ast.Ident)
	right, rightOK := assignment.Rhs[0].(*ast.Ident)
	return leftOK && rightOK && local != "_" && left.Name == "_" && right.Name == local
}
