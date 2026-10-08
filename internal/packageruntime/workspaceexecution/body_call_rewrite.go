package workspaceexecution

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func (c workspaceCalls) rewriteCalls(key string, d workspaceCallDeclaration, source, surface string, body bool) (string, []WorkspaceCallSite, error) {
	if source == "" {
		return source, nil, nil
	}
	set := token.NewFileSet()
	prefix := ""
	var root ast.Node
	var err error
	if body {
		prefix = "package scoped\nvar _ = func() {\n"
		root, err = parser.ParseFile(set, "computes", prefix+recordBodyGoSyntax(source)+"\n}", 0)
	} else {
		root, err = parser.ParseExprFrom(set, "expression", source, 0)
	}
	if err != nil {
		return "", nil, fmt.Errorf("parse call surface %s: %w", surface, err)
	}
	var sites []WorkspaceCallSite
	ast.Inspect(root, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || err != nil {
			return err == nil
		}
		if name, plain := call.Fun.(*ast.Ident); plain && name.Obj == nil && bodycodegen.IsBodyPrimitiveName(name.Name) {
			if _, declared := c[packageActivityKey(d.ref.PackagePath, name.Name)]; !declared {
				// Keep the primitive text and descend into its arguments so that
				// nested source activity calls still enter the dependency graph.
				return true
			}
		}
		var target string
		target, err = c.resolveCall(d, call)
		if err == nil {
			sites = append(sites, WorkspaceCallSite{key, target, surface,
				set.Position(call.Fun.Pos()).Offset - len(prefix), set.Position(call.Fun.End()).Offset - len(prefix)})
		}
		return err == nil
	})
	if err != nil {
		return "", nil, err
	}
	changes := append([]WorkspaceCallSite(nil), sites...)
	sort.Slice(changes, func(i, j int) bool { return changes[i].Start > changes[j].Start })
	for _, change := range changes {
		source = source[:change.Start] + c[change.Callee].ref.LoweredName + source[change.End:]
	}
	return source, sites, nil
}

func (c workspaceCalls) resolveCall(d workspaceCallDeclaration, call *ast.CallExpr) (string, error) {
	pkg, name := d.ref.PackagePath, ""
	var reference *ast.Ident
	switch function := call.Fun.(type) {
	case *ast.Ident:
		reference, name = function, function.Name
	case *ast.SelectorExpr:
		var ok bool
		reference, ok = function.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("call requires an activity name or one import alias")
		}
		pkg, name = "", function.Sel.Name
		for _, imported := range d.imports {
			alias := imported.Alias
			if alias == "" {
				alias = packagePathBase(imported.Path)
			}
			if alias == reference.Name {
				if pkg != "" {
					return "", fmt.Errorf("ambiguous call import alias %q", alias)
				}
				pkg = imported.Path
			}
		}
	default:
		return "", fmt.Errorf("call requires a source-declared named activity")
	}
	if reference.Obj != nil || isActivityInput(reference.Name, len(d.activity.Inputs)) {
		return "", fmt.Errorf("call name %q is shadowed by a local value", reference.Name)
	}
	key := packageActivityKey(pkg, name)
	callee, ok := c[key]
	if !ok || pkg == "" || (!callee.activity.ValueProgramPresent && callee.activity.Assembly == nil) {
		return "", fmt.Errorf("call %q requires a computes or assembling activity in the local package or a source import", name)
	}
	return key, nil
}

func isActivityInput(name string, count int) bool {
	if count == 1 {
		return name == "input"
	}
	for i := range count {
		if name == fmt.Sprintf("input%d", i) {
			return true
		}
	}
	return false
}
