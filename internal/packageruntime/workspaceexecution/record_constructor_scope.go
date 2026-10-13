package workspaceexecution

import (
	"fmt"
	"go/ast"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// Constructor imports belong to the activity's original file. Package-wide
// display names remain a separate lookup for the existing unqualified syntax.
type recordConstructorScope struct {
	types      map[string]string
	inputCount int
	body       string
}

func (n workspaceRecordNames) constructorScope(file *syntax.File, activity *syntax.ActivityDecl,
	image packageruntime.Image) recordConstructorScope {
	scope := recordConstructorScope{types: map[string]string{}, inputCount: len(activity.Inputs), body: activity.ValueProgram}
	if scope.body == "" && activity.Assembly != nil {
		scope.body = activity.Assembly.Spec.Baseline
	}
	for alias, path := range importAliases("", file.Imports) {
		for _, pkg := range image.Packages {
			if pkg.Path != path {
				continue
			}
			for _, export := range pkg.Exports {
				if export.Kind == "entity" && n.recordIDs[export.ID] {
					scope.types[alias+"."+export.Name] = n.canonical[export.ID]
				}
			}
		}
	}
	return scope
}

func (n workspaceRecordNames) constructorTypeName(pkg string, expression ast.Expr,
	scopes []recordConstructorScope) (string, error) {
	switch name := expression.(type) {
	case *ast.Ident:
		if !n.known[name.Name] {
			return "", nil
		}
		lowered := n.visible[pkg][name.Name]
		if lowered == "" {
			return "", fmt.Errorf("record constructor %q is undeclared or ambiguous in package %q", name.Name, pkg)
		}
		return lowered, nil
	case *ast.SelectorExpr:
		var scope recordConstructorScope
		if len(scopes) != 0 {
			scope = scopes[0]
		}
		return qualifiedConstructorName(name, scope)
	default:
		return "", nil
	}
}

func qualifiedConstructorName(expression *ast.SelectorExpr, scope recordConstructorScope) (string, error) {
	alias, ok := expression.X.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("record constructor requires one source import alias and entity member")
	}
	if alias.Obj != nil || isActivityInput(alias.Name, scope.inputCount) {
		return "", fmt.Errorf("record constructor alias %q is shadowed by a local value", alias.Name)
	}
	spelling := alias.Name + "." + expression.Sel.Name
	lowered := scope.types[spelling]
	if lowered == "" {
		return "", fmt.Errorf("record constructor %q requires a record exported by an import in this source file", spelling)
	}
	return lowered, nil
}
