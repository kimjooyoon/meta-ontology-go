package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// BodySourceReference records syntax, including calls through local values.
// Resolution, reachability and execution are separate observations.
type BodySourceReference struct {
	Name           string
	Direct         bool
	Ordinal        int
	AmbiguousScope bool
}

// InspectBodySourceReferences reuses the body's let normalization. It performs
// no projection, candidate selection, type checking, model call or execution.
// Enclosing bodies are optional for isolated candidate expressions. A name
// declared anywhere in those bodies makes that fragment call's scope ambiguous;
// candidate materialization must locate the precise enclosing lexical scope.
func InspectBodySourceReferences(body string, inputCount int, enclosing ...string) ([]BodySourceReference, error) {
	file, err := parseBodyReferenceFile(body, inputCount)
	if err != nil {
		return nil, err
	}
	bound, err := enclosingReferenceBindings(enclosing, inputCount)
	if err != nil {
		return nil, err
	}
	refs := []BodySourceReference{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ref := BodySourceReference{Name: sourceReferenceName(call.Fun), Ordinal: len(refs)}
		if ident, ok := ast.Unparen(call.Fun).(*ast.Ident); ok {
			ref.Direct = ident.Obj == nil
			if ref.Direct && bound[ident.Name] {
				ref.Direct, ref.AmbiguousScope = false, true
			}
		}
		refs = append(refs, ref)
		return true
	})
	return refs, nil
}

func parseBodyReferenceFile(body string, inputCount int) (*ast.File, error) {
	if inputCount < 0 || inputCount > 16 {
		return nil, fmt.Errorf("body input count %d is outside 0..16", inputCount)
	}
	parameters := make([]string, inputCount)
	for i := range parameters {
		name := "input"
		if inputCount > 1 {
			name = fmt.Sprintf("input%d", i)
		}
		parameters[i] = name + " any"
	}
	lowered, err := rewriteLetDeclarations(body)
	if err != nil {
		return nil, err
	}
	header := "package p\nvar _ = func(" + strings.Join(parameters, ",") + "){\n"
	return parser.ParseFile(token.NewFileSet(), "references.go", header+lowered+"\n}", 0)
}

func enclosingReferenceBindings(bodies []string, inputCount int) (map[string]bool, error) {
	bound := map[string]bool{}
	for _, body := range bodies {
		if body == "" {
			continue
		}
		file, err := parseBodyReferenceFile(body, inputCount)
		if err != nil {
			return nil, fmt.Errorf("enclosing body: %w", err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			if ident, ok := n.(*ast.Ident); ok && ident.Obj != nil {
				bound[ident.Name] = true
			}
			return true
		})
	}
	return bound, nil
}

func sourceReferenceName(expression ast.Expr) string {
	switch expr := ast.Unparen(expression).(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		return sourceReferenceName(expr.X) + "." + expr.Sel.Name
	default:
		return "<expression>"
	}
}
