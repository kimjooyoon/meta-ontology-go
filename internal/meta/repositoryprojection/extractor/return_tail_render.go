package extractor

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

func renderReturnTailHelper(fset *token.FileSet, name string, bindings []suffixBinding, body []byte) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString("func ")
	output.WriteString(name)
	output.WriteByte('(')
	for index, binding := range bindings {
		if index > 0 {
			output.WriteString(", ")
		}
		output.WriteString(binding.name)
		output.WriteByte(' ')
		if err := format.Node(&output, fset, binding.type_); err != nil {
			return nil, fail("derive-recipe", "render-return-tail-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
		}
	}
	output.WriteString(") error {\n")
	output.Write(body)
	if len(body) == 0 || body[len(body)-1] != '\n' {
		output.WriteByte('\n')
	}
	output.WriteString("}\n")
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, fail("derive-recipe", "render-return-tail-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
	}
	return formatted, nil
}

func renderedFunctionHelper(source []byte, name string) ([]byte, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "decomposed.go", source, parser.ParseComments)
	if err != nil {
		return nil, returnTailContradiction(obligationRenderedCapacity, "rendered source is not parseable")
	}
	list, err := imports(file)
	if err != nil {
		return nil, err
	}
	for _, node := range file.Decls {
		function, ok := node.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Name == nil || function.Name.Name != name {
			continue
		}
		start := function.Pos()
		if function.Doc != nil {
			start = function.Doc.Pos()
		}
		selected := []declaration{{node: function, start: fset.Position(start).Offset, end: fset.Position(function.End()).Offset, identity: "func:" + name}}
		_, helperImports, renderErr := renderImports(fset, file, []ast.Decl{function}, list, false)
		if renderErr != nil {
			return nil, renderErr
		}
		helper, err := renderSelectedHelper(fset, file, source, selected, helperImports)
		if err != nil {
			return nil, fail("generate-helpers", "format-helper", "AST_RENDER_FAILED", "DIRECT_MISSING", "restore-parser-evidence", nil)
		}
		return helper, nil
	}
	return nil, returnTailContradiction(obligationRenderedCapacity, "rendered target function is missing")
}

func renderedFunctionExceedsLimit(source []byte, name string) bool {
	rendered, err := renderedFunctionHelper(source, name)
	return err == nil && physicalLines(rendered) > functionLineLimit
}

func renderedCapacityProgress(before, after []byte) bool {
	beforeLines, afterLines := physicalLines(before), physicalLines(after)
	return afterLines < beforeLines || beforeLines == afterLines && len(after) < len(before)
}
