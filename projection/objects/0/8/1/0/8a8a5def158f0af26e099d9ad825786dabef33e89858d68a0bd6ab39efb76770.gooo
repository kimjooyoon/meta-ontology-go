package analyzer

import (
	"go/ast"
	"go/token"
	"sort"
)

func relationForReference(entries []Registration) Relation {
	if allOfKind(entries, KindEntity) {
		return RelationUses
	}
	return RelationReferences
}
func allOfKind(entries []Registration, kind SymbolKind) bool {
	if len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if entry.Kind != kind {
			return false
		}
	}
	return true
}
func uniqueIdentities(entries []Registration) []Identity {
	seen := make(map[Identity]bool, len(entries))
	options := make([]Identity, 0, len(entries))
	for _, entry := range entries {
		if entry.Identity.Valid() && !seen[entry.Identity] {
			seen[entry.Identity] = true
			options = append(options, entry.Identity)
		}
	}
	sort.Slice(options, func(i, j int) bool { return identityLess(options[i], options[j]) })
	return options
}
func localBindings(function *ast.FuncDecl) map[string]bool {
	blocked := make(map[string]bool)
	collectFieldNames(blocked, function.Recv)
	collectFieldNames(blocked, function.Type.Params)
	collectFieldNames(blocked, function.Type.Results)
	if function.Body == nil {
		return blocked
	}
	ast.Inspect(function.Body, func(node ast.Node) bool {
		switch current := node.(type) {
		case *ast.ValueSpec:
			addNames(blocked, current.Names)
		case *ast.AssignStmt:
			if current.Tok == token.DEFINE {
				addExprNames(blocked, current.Lhs)
			}
		case *ast.RangeStmt:
			addExprName(blocked, current.Key)
			addExprName(blocked, current.Value)
		case *ast.TypeSpec:
			blocked[current.Name.Name] = true
		case *ast.FuncLit:
			collectFieldNames(blocked, current.Type.Params)
			collectFieldNames(blocked, current.Type.Results)
		}
		return true
	})
	return blocked
}
func collectFieldNames(blocked map[string]bool, fields *ast.FieldList) {
	if fields == nil {
		return
	}
	for _, field := range fields.List {
		addNames(blocked, field.Names)
	}
}
