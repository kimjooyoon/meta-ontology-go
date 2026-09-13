package extractor

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"strconv"
	"strings"
)

func renderMapLiteralCandidate(source []byte, fset *token.FileSet, file *ast.File, function *ast.FuncDecl, literal *ast.CompositeLit, keys []string, pkg *types.Package) (*suffixCandidate, error) {
	start, end := fset.Position(literal.Pos()).Offset, fset.Position(literal.End()).Offset
	if start < 0 || end > len(source) || start >= end {
		return nil, returnTailContradiction(obligationControlFlow, "map literal source range is invalid")
	}
	name := mapLiteralHelperName(file, pkg, function.Name.Name)
	helper, call, err := renderMapLiteralParts(source, fset, literal, keys, name)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	output.Write(source[:start])
	output.WriteString(call)
	output.Write(source[end:])
	output.WriteString("\n")
	output.Write(helper)
	combined, err := format.Source(output.Bytes())
	if err != nil {
		return nil, returnTailContradiction(obligationProjectedConform, "map construction candidate did not render")
	}
	before, err := mapLiteralMeasurement(source, function.Name.Name)
	if err != nil {
		return nil, err
	}
	outer, err := mapLiteralMeasurement(combined, function.Name.Name)
	if err != nil {
		return nil, err
	}
	inner, err := mapLiteralMeasurement(combined, name)
	if err != nil {
		return nil, err
	}
	after := outer.overage + inner.overage
	if !renderedCapacityProgress(renderedCapacitySnapshot{overage: before.overage}, renderedCapacitySnapshot{overage: after}) {
		return nil, returnTailContradiction(obligationRenderedCapacity, "map construction does not strictly reduce total rendered overage")
	}
	return &suffixCandidate{
		start: start, end: end, helperName: name, helper: helper, result: combined,
		beforeRenderedCapacityOverage: before.overage, afterRenderedCapacityOverage: after,
		renderedHelper: inner, renderedOuter: outer,
	}, nil
}

func renderMapLiteralParts(source []byte, fset *token.FileSet, literal *ast.CompositeLit, keys []string, name string) ([]byte, string, error) {
	parameters := make([]string, 0, len(keys))
	arguments := make([]string, 0, len(keys))
	var body strings.Builder
	for index, key := range keys {
		pair := literal.Elts[index].(*ast.KeyValueExpr)
		start, end := fset.Position(pair.Value.Pos()).Offset, fset.Position(pair.Value.End()).Offset
		if start < 0 || end > len(source) || start >= end {
			return nil, "", returnTailContradiction(obligationFreeBindings, "map value source range is invalid")
		}
		parameter := fmt.Sprintf("value%d", index)
		parameters = append(parameters, parameter+" any")
		arguments = append(arguments, string(source[start:end]))
		fmt.Fprintf(&body, "\t\t%s: %s,\n", strconv.Quote(key), parameter)
	}
	helper := "func " + name + "(" + strings.Join(parameters, ", ") + ") map[string]any {\n\treturn map[string]any{\n" + body.String() + "\t}\n}\n"
	formatted, err := format.Source([]byte(helper))
	if err != nil {
		return nil, "", returnTailContradiction(obligationProjectedConform, "map constructor did not render")
	}
	return formatted, name + "(" + strings.Join(arguments, ", ") + ")", nil
}

func mapLiteralMeasurement(source []byte, name string) (renderedCapacityMeasurement, error) {
	rendered, err := renderedFunctionHelper(source, name)
	if err != nil {
		return renderedCapacityMeasurement{}, err
	}
	return canonicalRenderedCapacity(rendered)
}
