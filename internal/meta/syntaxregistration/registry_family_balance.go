package syntaxregistration

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Preserve the already authored family width rather than adding another
// historical size constant. Only direct, comment-free data constructors can be
// repartitioned; unresolved source layouts retain the ordinary append path.
func (source *goSource) appendBoundedRegistryFamily(root *ast.FuncDecl, tail *ast.CompositeLit, row string) (bool, error) {
	leaves, err := source.registryCaseLeaves(root)
	if err != nil {
		return false, err
	}
	if len(leaves) < 2 {
		return false, nil
	}
	width := 0
	for _, leaf := range leaves[:len(leaves)-1] {
		if len(leaf.Elts) > 0 && (width == 0 || len(leaf.Elts) < width) {
			width = len(leaf.Elts)
		}
	}
	if width < 1 || len(tail.Elts) < width {
		return false, nil
	}
	owner := source.registryLiteralOwner(tail)
	if owner == nil {
		return false, nil
	}
	parsed, err := parser.ParseFile(token.NewFileSet(), "family.go", "package owned\n"+source.text(owner), parser.ParseComments)
	if err != nil {
		return false, err
	}
	if len(parsed.Comments) != 0 {
		return false, fmt.Errorf("bounded registry family requires a comment-preserving source split")
	}
	rows := make([]string, 0, len(tail.Elts)+1)
	for _, element := range tail.Elts {
		rows = append(rows, source.text(element))
	}
	rows = append(rows, strings.TrimSuffix(strings.TrimSpace(row), ","))
	source.replace(owner, source.renderRegistryFamily(owner.Name.Name, rows, width))
	return true, nil
}

func (source *goSource) registryLiteralOwner(list *ast.CompositeLit) *ast.FuncDecl {
	for _, declaration := range source.file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || function.Body == nil || len(function.Body.List) != 1 {
			continue
		}
		returned, ok := function.Body.List[0].(*ast.ReturnStmt)
		if ok && len(returned.Results) == 1 && returned.Results[0] == list {
			return function
		}
	}
	return nil
}

func (source *goSource) renderRegistryFamily(name string, rows []string, width int) string {
	used := map[string]bool{}
	ast.Inspect(source.file, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			used[id.Name] = true
		}
		return true
	})
	var names []string
	for len(names) < (len(rows)+width-1)/width {
		for next := 0; ; next++ {
			candidate := fmt.Sprintf("%sPart%d", name, next)
			if !used[candidate] {
				names = append(names, candidate)
				used[candidate] = true
				break
			}
		}
	}
	var result strings.Builder
	fmt.Fprintf(&result, "func %s() []CaseDefinition {\ncases := %s()\n", name, names[0])
	for _, child := range names[1:] {
		fmt.Fprintf(&result, "cases = append(cases, %s()...)\n", child)
	}
	result.WriteString("return cases\n}\n")
	for i, child := range names {
		fmt.Fprintf(&result, "func %s() []CaseDefinition {\nreturn []CaseDefinition{\n", child)
		for _, row := range rows[i*width : min((i+1)*width, len(rows))] {
			result.WriteString(row + ",\n")
		}
		result.WriteString("}\n}\n")
	}
	return result.String()
}
