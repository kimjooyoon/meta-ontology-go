package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

// FieldBinding identifies a source local or input, independently of its values.
// Input ordinals are negative and local ordinals retain existing source identity.
type CompositionFieldBinding struct {
	RootActivityID    string                            `json:"root_activity_id"`
	ActivityID        string                            `json:"activity_id"`
	VariableOrdinal   int                               `json:"variable_ordinal"`
	Name              string                            `json:"name"`
	InputPort         string                            `json:"input_port,omitempty"`
	InputKind         string                            `json:"input_kind,omitempty"`
	InputFrom         string                            `json:"input_from_activity_id,omitempty"`
	Record            bodycodegen.RecordType            `json:"record"`
	Declaration       *bodycodegen.BodyExpressionSource `json:"declaration,omitempty"`
	ExpressionKind    string                            `json:"expression_kind,omitempty"`
	ExpressionOrdinal int                               `json:"expression_ordinal,omitempty"`
}

type compositionFieldInventory struct {
	bindings []CompositionFieldBinding
	locals   map[conditionOriginKey]int
	inputs   map[conditionOriginKey]int
	returns  map[conditionOriginKey]int
}

// Resolve source identities against the original typed Go projection. This
// inventory contains no runtime values and does not infer identity by equality.
func compositionFieldBindings(prior Composition) (compositionFieldInventory, error) {
	r := compositionFieldInventory{locals: map[conditionOriginKey]int{}, inputs: map[conditionOriginKey]int{}, returns: map[conditionOriginKey]int{}}
	records := map[string]bodycodegen.RecordType{}
	for _, step := range append(append([]CompositionStep(nil), prior.Steps...), prior.Preparations...) {
		for _, record := range step.Generation.Report.RecordTypes {
			eligible := true
			for _, field := range record.Fields {
				_, scalar := compositionRecordFieldScalar(field.TypeID)
				eligible = eligible && scalar && field.Presence == "" && field.Cardinality == ""
			}
			if eligible && len(record.Fields) > 0 && len(record.Fields) <= bodycodegen.RecordFieldLimit {
				records[record.GoName] = record
			}
		}
	}
	if len(records) == 0 {
		return r, nil
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "field-inventory.go", prior.Source, 0)
	if err != nil {
		return r, err
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	if _, err := new(types.Config).Check("gooo.field.inventory", fset, []*ast.File{file}, info); err != nil {
		return r, err
	}
	origins, err := compositionSourceLocals(prior)
	if err != nil {
		return r, err
	}
	returnSources, err := compositionSourceReturns(prior)
	if err != nil {
		return r, err
	}
	functions := arithmeticFunctionIDs(prior)
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || functions[fn.Name.Name][0] == "" {
			continue
		}
		ids := functions[fn.Name.Name]
		objects := r.addParameters(fn, ids, info, records, prior)
		ordinal := r.addLocals(fn, ids, info, records, objects, origins)
		if err := r.addReturns(fn, ids, info, records, objects, ordinal, returnSources); err != nil {
			return r, err
		}
	}
	if len(r.bindings) > 4096 {
		return r, fmt.Errorf("record field binding inventory exceeds4096")
	}
	return r, nil
}

func fieldObjectRecord(object types.Object, records map[string]bodycodegen.RecordType) (bodycodegen.RecordType, bool) {
	if object != nil {
		if named, ok := object.Type().(*types.Named); ok {
			record, found := records[named.Obj().Name()]
			return record, found
		}
	}
	return bodycodegen.RecordType{}, false
}

func fieldInputKind(prior Composition, ids [2]string, port int) (string, string) {
	if ids[0] == ids[1] {
		for _, node := range prior.Plan.Activities {
			if node.ID == ids[0] {
				slots := node.inputSlots()
				if port < len(slots) {
					if slots[port].From < 0 {
						return "root_input", ""
					}
					if slots[port].From < len(prior.Plan.Activities) {
						return "activity_input", prior.Plan.Activities[slots[port].From].ID
					}
				}
			}
		}
	}
	return "helper_parameter", ""
}
