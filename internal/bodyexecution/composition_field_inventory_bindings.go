package bodyexecution

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
	"go/ast"
	"go/types"
)

func (r *compositionFieldInventory) addParameters(fn *ast.FuncDecl, ids [2]string, info *types.Info, records map[string]bodycodegen.RecordType, prior Composition) map[types.Object]int {
	objects := map[types.Object]int{}
	port := 0
	for _, group := range fn.Type.Params.List {
		for _, name := range group.Names {
			if record, ok := fieldObjectRecord(info.ObjectOf(name), records); ok {
				key := conditionOriginKey{ids[0], ids[1], -port - 1}
				r.inputs[key] = len(r.bindings)
				objects[info.ObjectOf(name)] = len(r.bindings)
				kind, from := fieldInputKind(prior, ids, port)
				r.bindings = append(r.bindings, CompositionFieldBinding{RootActivityID: ids[0], ActivityID: ids[1],
					VariableOrdinal: key.Ordinal, Name: name.Name, InputPort: name.Name, InputKind: kind, InputFrom: from, Record: record})
			}
			port++
		}
	}
	return objects
}

func (r *compositionFieldInventory) addLocals(fn *ast.FuncDecl, ids [2]string, info *types.Info, records map[string]bodycodegen.RecordType, objects map[types.Object]int, origins map[conditionOriginKey]compositionLocalOrigin) int {
	ordinal := 0
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		name, _, _, _ := bodycodegen.LocalWriteExpression(n)
		if name == nil {
			return true
		}
		key := conditionOriginKey{ids[0], ids[1], ordinal}
		ordinal++
		origin, ok := origins[key]
		if !ok {
			return true
		}
		object := info.ObjectOf(name)
		record, ok := fieldObjectRecord(object, records)
		if !ok {
			return true
		}
		index, found := objects[object]
		if !found {
			index = len(r.bindings)
			objects[object] = index
			r.bindings = append(r.bindings, CompositionFieldBinding{RootActivityID: ids[0], ActivityID: ids[1],
				VariableOrdinal: origin.definition.VariableOrdinal, Name: name.Name, Record: record, Declaration: origin.source.Declaration})
		}
		r.locals[key] = index
		return true
	})
	return ordinal
}
