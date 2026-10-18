package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
)

type collectionMapperBinding struct {
	Name, Input, Output string
}

func (r *pureCallResolver) collectionMapper(call *ast.CallExpr, caller string) (collectionMapperBinding, error) {
	var binding collectionMapperBinding
	if call == nil || len(call.Args) != 2 {
		return binding, fmt.Errorf("map_values requires a collection and a named unary scalar activity")
	}
	mapper, ok := call.Args[1].(*ast.Ident)
	if !ok || r.ids[mapper.Name] == "" {
		return binding, fmt.Errorf("map_values mapper must name a source-declared activity")
	}
	owner, err := sourceBodyActivity(r.file, caller)
	if err != nil {
		return binding, err
	}
	inputs, err := sourceBodyParameters(owner, r.records, r.file)
	if err != nil {
		return binding, err
	}
	if err := r.mapNameInSource(call.Fun.(*ast.Ident), "call", caller, inputs); err != nil {
		return binding, err
	}
	if err := r.mapNameInSource(mapper, "mapper", caller, inputs); err != nil {
		return binding, err
	}
	activity, err := sourceBodyActivity(r.file, mapper.Name)
	if err != nil || activity.Assembly != nil {
		return binding, fmt.Errorf("map_values mapper requires a fixed computes activity")
	}
	parameters, err := sourceBodyParameters(activity, r.records, r.file)
	output, found := bodyEntityType(activity.Output, r.records, r.file)
	scalar := func(t string) bool { return t == "int64" || t == "string" || t == "bool" }
	if err != nil || len(parameters) != 1 || !scalar(parameters[0].Type) || !found || !scalar(output) {
		return binding, fmt.Errorf("map_values mapper requires one Integer, Text or Boolean input and scalar output")
	}
	return collectionMapperBinding{mapper.Name, parameters[0].Type, output}, nil
}

func (r *pureCallResolver) mapNameInSource(name *ast.Ident, role, caller string, inputs []InputParameter) error {
	scope := ""
	if name.Obj != nil {
		scope = "local declaration"
	} else {
		for _, input := range inputs {
			if input.Name == name.Name {
				scope = "input"
			}
		}
	}
	if scope != "" {
		return fmt.Errorf("map_values %s %q is shadowed by %s in activity %q (%s)", role, name.Name, scope, caller, r.ids[caller])
	}
	return nil
}

// A typed local function keeps the primitive within its owning source activity.
// Its collection argument evaluates the original collection exactly once. The output
// backing array uses the validated collection limit; the result caps capacity to its length.
// Passing the mapper outside this scope avoids capturing source activity names.
func lowerCollectionMap(call *ast.CallExpr, binding collectionMapperBinding) error {
	text := fmt.Sprintf(`func(goooMapInput []%s, goooMapApply func(%s) %s) []%s {
 var goooMapResult [%d]%s
 var goooMapLength int
 for goooMapIndex := range goooMapInput {
  goooMapResult[goooMapIndex] = goooMapApply(goooMapInput[goooMapIndex])
  goooMapLength = goooMapIndex + 1
 }
 return goooMapResult[:goooMapLength:goooMapLength]
}(goooMapArgument, goooMapMapper)`, binding.Input, binding.Input, binding.Output, binding.Output, RepeatedFieldLimit, binding.Output)
	expression, err := parser.ParseExpr(text)
	if err != nil {
		return err
	}
	projected := expression.(*ast.CallExpr)
	projected.Args = call.Args
	*call = *projected
	return nil
}

func (p bodyProjection) lowerCollectionMaps() error {
	if _, shadowed := findFunction(p.file, "map_values"); shadowed {
		return nil
	}
	bindings := map[string]collectionMapperBinding{}
	for _, f := range p.calls {
		if len(f.parameters) == 1 {
			bindings[f.identity.Name] = collectionMapperBinding{f.identity.Name, f.parameters[0].Type, f.output}
		}
	}
	var err error
	ast.Inspect(p.file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || err != nil {
			return err == nil
		}
		name, ok := call.Fun.(*ast.Ident)
		if !ok || name.Name != "map_values" {
			return true
		}
		if len(call.Args) != 2 {
			err = fmt.Errorf("map_values requires two arguments")
			return false
		}
		mapper, ok := call.Args[1].(*ast.Ident)
		if !ok || bindings[mapper.Name].Name == "" {
			err = fmt.Errorf("map_values mapper has no resolved source binding")
			return false
		}
		err = lowerCollectionMap(call, bindings[mapper.Name])
		return err == nil
	})
	return err
}
