package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func (r *compositionFieldInventory) addReturns(function *ast.FuncDecl, ids [2]string,
	info *types.Info, records map[string]bodycodegen.RecordType, objects map[types.Object]int,
	localSlots int, origins map[conditionOriginKey]compositionReturnSource) error {
	statements := fieldFunctionReturns(function)
	for ordinal, statement := range statements {
		key := conditionOriginKey{ids[0], ids[1], ordinal}
		retained, observed := origins[key]
		if !observed {
			continue
		}
		if retained.origin.NativePopulation != len(statements) {
			return fmt.Errorf("record return source population differs")
		}
		expression := unparenthesizedFieldValue(statement.Results[0])
		if name, ok := expression.(*ast.Ident); ok {
			if _, bound := objects[info.ObjectOf(name)]; bound {
				continue
			}
		}
		named, ok := info.TypeOf(expression).(*types.Named)
		if !ok {
			return fmt.Errorf("record return has no named native type")
		}
		record, eligible := records[named.Obj().Name()]
		if !eligible {
			continue
		}
		if retained.origin.RecordID != record.ID {
			return fmt.Errorf("record return source type identity differs")
		}
		r.returns[key] = len(r.bindings)
		r.bindings = append(r.bindings, CompositionFieldBinding{RootActivityID: ids[0], ActivityID: ids[1],
			VariableOrdinal: localSlots + ordinal, Name: fmt.Sprintf("<return:%d>", ordinal), Record: record,
			Declaration: retained.source, ExpressionKind: "return", ExpressionOrdinal: ordinal})
	}
	return nil
}
