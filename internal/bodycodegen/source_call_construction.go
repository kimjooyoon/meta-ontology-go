package bodycodegen

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

// CallConstruction lists called assembling activities in dependency order.
// Deferred names need those bodies before their complete body can be checked.
type CallConstruction struct {
	Helpers  []string
	Deferred []string
}

func PlanCallConstruction(ctx context.Context, filename string, source []byte, roots []string) (CallConstruction, error) {
	var result CallConstruction
	if ctx == nil || ctx.Err() != nil {
		return result, fmt.Errorf("call construction requires an active context")
	}
	file, diagnostics := ParseBodyFile(filename, source)
	if diagnostics.HasErrors() || file == nil {
		return result, fmt.Errorf("call construction source has syntax errors")
	}
	p := callConstructionPlanner{activities: map[string]*syntax.ActivityDecl{}, heights: map[string]int{},
		visiting: map[string]bool{}, visited: map[string]bool{}, assemblingBelow: map[string]bool{}, called: map[string]bool{}}
	for _, declaration := range file.Declarations {
		if activity, ok := declaration.(*syntax.ActivityDecl); ok {
			p.activities[activity.Name] = activity
		}
	}
	for _, root := range roots {
		if err := p.visit(root, 0); err != nil {
			return result, err
		}
	}
	for _, name := range p.order {
		if p.called[name] && p.activities[name].Assembly != nil {
			result.Helpers = append(result.Helpers, name)
		}
		if p.assemblingBelow[name] {
			result.Deferred = append(result.Deferred, name)
		}
	}
	if len(result.Helpers) > 32 {
		return result, fmt.Errorf("call construction exceeds 32 assembled helpers")
	}
	return result, nil
}

type callConstructionPlanner struct {
	activities      map[string]*syntax.ActivityDecl
	heights         map[string]int
	visiting        map[string]bool
	visited         map[string]bool
	assemblingBelow map[string]bool
	called          map[string]bool
	order           []string
}

func (p *callConstructionPlanner) visit(name string, depth int) error {
	if p.visiting[name] {
		return fmt.Errorf("recursive pure activity call at %q", name)
	}
	if depth+p.heights[name] > 16 {
		return fmt.Errorf("pure activity closure exceeds 16 call levels at %q", name)
	}
	if p.visited[name] {
		return nil
	}
	activity := p.activities[name]
	if activity == nil {
		return fmt.Errorf("call construction root %q is undeclared", name)
	}
	p.visiting[name] = true
	calls, err := activityConstructionCalls(activity)
	if err != nil {
		return fmt.Errorf("activity %s call surfaces: %w", name, err)
	}
	for _, callee := range calls {
		declaration := p.activities[callee]
		if declaration == nil {
			// The body/choice checker retains unknown functions as type failures.
			continue
		}
		p.called[callee] = true
		if err := p.visit(callee, depth+1); err != nil {
			return err
		}
		p.assemblingBelow[name] = p.assemblingBelow[name] || declaration.Assembly != nil || p.assemblingBelow[callee]
		p.heights[name] = max(p.heights[name], 1+p.heights[callee])
	}
	p.visiting[name], p.visited[name] = false, true
	p.order = append(p.order, name)
	return nil
}

func activityConstructionCalls(activity *syntax.ActivityDecl) ([]string, error) {
	surfaces := []string{activity.ValueProgram}
	if activity.Assembly != nil {
		spec := activity.Assembly.Spec
		surfaces = append(surfaces, spec.Baseline)
		for _, choice := range spec.Choices {
			if choice.Kind == "field_value" || choice.Kind == "field_update" {
				surfaces = append(surfaces, "return "+choice.Alternative)
			}
		}
		if spec.FillPlan != nil {
			for _, candidate := range spec.FillPlan.Candidates {
				for _, fill := range candidate.Fills {
					surfaces = append(surfaces, "return "+fill.Expression)
				}
			}
		}
	}
	set := map[string]bool{}
	for _, surface := range surfaces {
		if surface == "" {
			continue
		}
		body, err := rewriteLetDeclarations(surface)
		if err != nil {
			return nil, err
		}
		file, err := parser.ParseFile(token.NewFileSet(), "calls.go", "package p\nvar _ = func(){\n"+body+"\n}", 0)
		if err != nil {
			return nil, err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok {
				if ident, ok := call.Fun.(*ast.Ident); ok && ident.Obj == nil {
					set[ident.Name] = true
				}
			}
			return true
		})
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}
