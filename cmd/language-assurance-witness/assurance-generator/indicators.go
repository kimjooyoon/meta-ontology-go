package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"os"
)

func transformIndicators(path string) ([]byte, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	files, file, err := parseSource(path, source)
	if err != nil {
		return nil, err
	}
	seamFound := false
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil || !hasIdentifier(function, "MetricRawReconstruction") {
			continue
		}
		parameter := firstParameter(function)
		ast.Inspect(function.Body, func(node ast.Node) bool {
			result, ok := node.(*ast.ReturnStmt)
			if !ok || len(result.Results) != 1 || !hasIdentifier(result.Results[0], "MetricRawReconstruction") {
				return true
			}
			seamFound = true
			if !hasIdentifier(result.Results[0], "writeSetIndicator") {
				result.Results[0] = &ast.CallExpr{Fun: ast.NewIdent("append"), Args: []ast.Expr{result.Results[0], &ast.CallExpr{Fun: ast.NewIdent("writeSetIndicator"), Args: []ast.Expr{ast.NewIdent(parameter)}}}}
			}
			return false
		})
	}
	if !seamFound {
		return nil, fmt.Errorf("indicator return seam not found")
	}
	return formatSource(files, file)
}

func firstParameter(function *ast.FuncDecl) string {
	for _, field := range function.Type.Params.List {
		if len(field.Names) > 0 {
			return field.Names[0].Name
		}
	}
	return "summary"
}

func renderIndicator(spec metricSpec) ([]byte, error) {
	source := fmt.Sprintf(`package languageassurance

func writeSetIndicator(summary Summary) Indicator {
	value := summary.RepositoryWrites
	result := indicator(%q, ClassGuardrail, ProofRegression, %q, &value, 0, %q, RelationLessOrEqual)
	result.Producer = %q
	result.Consumer = %q
	return result
}
`, spec.MetricID, spec.MetaOperation, spec.Unit, spec.Producer, spec.Consumer)
	return format.Source([]byte(source))
}
