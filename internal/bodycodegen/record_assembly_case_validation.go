package bodycodegen

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
)

// Preflight checks types and JSON shapes. It never evaluates a candidate or
// reveals finite outcomes before the optional prediction.
func validateRecordAssemblyCases(ctx context.Context, source []byte, activity string, records []RecordType, cases []assemblyspec.ValueCase) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "case-contract.go", source, parser.AllErrors)
	if err != nil {
		return err
	}
	function, ok := findFunction(file, activity)
	if !ok {
		return fmt.Errorf("record case function is missing")
	}
	information := types.Info{Types: make(map[ast.Expr]types.TypeAndValue)}
	if _, err = new(types.Config).Check(file.Name.Name, fset, []*ast.File{file}, &information); err != nil {
		return err
	}
	var inputs []types.Type
	for _, parameter := range function.Type.Params.List {
		for range parameter.Names {
			inputs = append(inputs, information.Types[parameter.Type].Type)
		}
	}
	output := information.Types[function.Type.Results.List[0].Type].Type
	for _, c := range cases {
		if err := ctx.Err(); err != nil {
			return err
		}
		var raw []json.RawMessage
		if err = json.Unmarshal([]byte(c.Inputs), &raw); err != nil || len(raw) != len(inputs) {
			return fmt.Errorf("value_case input must have exactly %d positional values", len(inputs))
		}
		for i, t := range inputs {
			if _, err = decodeRecordCaseValue(raw[i], t, records); err != nil {
				return fmt.Errorf("value_case input %d: %w", i, err)
			}
		}
		if _, err = decodeRecordCaseValue([]byte(c.Expected), output, records); err != nil {
			return fmt.Errorf("value_case expected: %w", err)
		}
	}
	return nil
}
