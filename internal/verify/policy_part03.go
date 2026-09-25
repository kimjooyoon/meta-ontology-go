package verify

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func checkSourceFile(root, path string, policy LinePolicy) []Violation {
	if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".gooo") {
		return nil
	}
	source, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return []Violation{{Path: path, Rule: "read-gooo-file", Detail: err.Error()}}
	}
	violations := make([]Violation, 0)
	if lines := lineCount(source); lines > policy.MaxFileLines {
		rule := "DAMP file lines"
		if strings.HasSuffix(path, ".gooo") {
			rule = "GOOO file lines"
		}
		violations = append(violations, Violation{Path: path, Rule: rule, Actual: lines, Limit: policy.MaxFileLines})
	}
	if strings.HasSuffix(path, ".gooo") {
		return violations
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return append(violations, Violation{Path: path, Rule: "parse-go-file", Detail: err.Error()})
	}
	ast.Inspect(file, func(node ast.Node) bool {
		start, end, name, ok := functionRange(fset, node)
		if !ok {
			return true
		}
		if length := end - start + 1; length > policy.MaxFunctionLines {
			violations = append(violations, Violation{Path: path, Rule: "DRY function lines", Actual: length, Limit: policy.MaxFunctionLines, Detail: name})
		}
		return true
	})
	return violations
}
func functionRange(fset *token.FileSet, node ast.Node) (int, int, string, bool) {
	var name string
	switch function := node.(type) {
	case *ast.FuncDecl:
		name = function.Name.Name
		if function.Recv != nil {
			name = "method " + name
		}
	case *ast.FuncLit:
		name = "function literal"
	default:
		return 0, 0, "", false
	}
	return fset.Position(node.Pos()).Line, fset.Position(node.End()).Line, name, true
}
