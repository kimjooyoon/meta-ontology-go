package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
)

func (r *pureCallResolver) collectionPredicate(call *ast.CallExpr, caller string) (collectionMapperBinding, error) {
	var binding collectionMapperBinding
	if call == nil || len(call.Args) != 2 {
		return binding, fmt.Errorf("filter_values requires a collection and a named unary predicate")
	}
	predicate, ok := call.Args[1].(*ast.Ident)
	if !ok || r.ids[predicate.Name] == "" {
		return binding, fmt.Errorf("filter_values predicate must name a source-declared activity")
	}
	owner, err := sourceBodyActivity(r.file, caller)
	if err != nil {
		return binding, err
	}
	inputs, err := sourceBodyParameters(owner, r.records, r.file)
	if err != nil {
		return binding, err
	}
	if err := r.collectionNameInSource("filter_values", call.Fun.(*ast.Ident), "call", caller, inputs); err != nil {
		return binding, err
	}
	if err := r.collectionNameInSource("filter_values", predicate, "predicate", caller, inputs); err != nil {
		return binding, err
	}
	activity, err := sourceBodyActivity(r.file, predicate.Name)
	if err != nil || activity.Assembly != nil {
		return binding, fmt.Errorf("filter_values predicate requires a fixed computes activity")
	}
	parameters, err := sourceBodyParameters(activity, r.records, r.file)
	output, found := bodyEntityType(activity.Output, r.records, r.file)
	scalar := func(t string) bool { return t == "int64" || t == "string" || t == "bool" }
	if err != nil || len(parameters) != 1 || !scalar(parameters[0].Type) || !found || output != "bool" {
		return binding, fmt.Errorf("filter_values predicate requires one Integer, Text or Boolean input and Boolean output")
	}
	return collectionMapperBinding{predicate.Name, parameters[0].Type, output}, nil
}

// The source collection and each predicate call are evaluated once. An array
// holds at most the declared 256 values, and the returned slice cannot grow.
func lowerCollectionFilter(call *ast.CallExpr, binding collectionMapperBinding) error {
	text := fmt.Sprintf(`func(goooFilterInput []%s, goooFilterApply func(%s) bool) []%s {
 var goooFilterResult [%d]%s
 var goooFilterLength int
 for goooFilterIndex := range goooFilterInput {
  if goooFilterApply(goooFilterInput[goooFilterIndex]) {
   goooFilterResult[goooFilterLength] = goooFilterInput[goooFilterIndex]
   goooFilterLength++
  }
 }
 return goooFilterResult[:goooFilterLength:goooFilterLength]
}(goooFilterArgument, goooFilterPredicate)`, binding.Input, binding.Input, binding.Input, RepeatedFieldLimit, binding.Input)
	expression, err := parser.ParseExpr(text)
	if err != nil {
		return err
	}
	projected := expression.(*ast.CallExpr)
	projected.Args = call.Args
	*call = *projected
	return nil
}

func (p bodyProjection) lowerCollectionFilters() error {
	if _, shadowed := findFunction(p.file, "filter_values"); shadowed {
		return nil
	}
	bindings := map[string]collectionMapperBinding{}
	for _, f := range p.calls {
		if len(f.parameters) == 1 && f.output == "bool" {
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
		if !ok || name.Name != "filter_values" {
			return true
		}
		if len(call.Args) != 2 {
			err = fmt.Errorf("filter_values requires two arguments")
			return false
		}
		predicate, ok := call.Args[1].(*ast.Ident)
		if !ok || bindings[predicate.Name].Name == "" {
			err = fmt.Errorf("filter_values predicate has no resolved source binding")
			return false
		}
		err = lowerCollectionFilter(call, bindings[predicate.Name])
		return err == nil
	})
	return err
}
