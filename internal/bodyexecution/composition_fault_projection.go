package bodyexecution

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"
)

type arithmeticProjectionSite struct {
	CompositionFaultSite
	node *ast.BinaryExpr
}

type arithmeticProjection struct {
	source string
	prefix string
	fset   *token.FileSet
	sites  []arithmeticProjectionSite
}

func arithmeticFunctionIDs(prior Composition) map[string][2]string {
	names := map[string][2]string{}
	for i, step := range prior.Steps {
		root := step.Generation.Report.ActivityID
		names[fmt.Sprintf("GoooComposedActivity%d", i)] = [2]string{root, root}
		if closure := step.Generation.Report.CallClosure; closure != nil {
			for c, helper := range closure.Activities {
				names[fmt.Sprintf("GoooComposedActivity%dCall%d", i, c)] = [2]string{root, helper.ActivityID}
			}
		}
	}
	return names
}

func arithmeticCompositionProjection(prior Composition) (string, []CompositionFaultSite, error) {
	p := arithmeticProjection{source: prior.Source, prefix: arithmeticObservationPrefix(prior.Source), fset: token.NewFileSet()}
	file, err := parser.ParseFile(p.fset, "composition.go", prior.Source, 0)
	if err != nil {
		return "", nil, err
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
	if _, err := new(types.Config).Check("gooo.observed.arithmetic", p.fset, []*ast.File{file}, info); err != nil {
		return "", nil, err
	}
	names := arithmeticFunctionIDs(prior)
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok {
			p.collect(function, names[function.Name.Name], info)
		}
	}
	if len(p.sites) == 0 {
		return prior.Source, nil, nil
	}
	slices.SortFunc(p.sites, func(a, b arithmeticProjectionSite) int {
		if a.Start == b.Start {
			return b.End - a.End
		}
		return a.Start - b.Start
	})
	sites := make([]CompositionFaultSite, len(p.sites))
	for i, site := range p.sites {
		if site.ActivityID == "" || site.RootActivityID == "" {
			return "", nil, fmt.Errorf("arithmetic operation has no source activity identity")
		}
		sites[i] = site.CompositionFaultSite
	}
	raw, err := format.Source([]byte(p.render(0, len(p.source))))
	return string(raw), sites, err
}

func (p *arithmeticProjection) collect(function *ast.FuncDecl, ids [2]string, info *types.Info) {
	ast.Inspect(function.Body, func(node ast.Node) bool {
		expression, ok := node.(*ast.BinaryExpr)
		if !ok || expression.Op != token.QUO && expression.Op != token.REM {
			return true
		}
		typed := info.Types[expression]
		if typed.Value != nil || typed.Type == nil || !types.Identical(typed.Type, types.Typ[types.Int64]) {
			return true
		}
		start, end := p.fset.Position(expression.Pos()).Offset, p.fset.Position(expression.End()).Offset
		site := CompositionFaultSite{RootActivityID: ids[0], ActivityID: ids[1], Operator: expression.Op.String(),
			Expression: p.source[start:end], ProjectionSHA256: digest([]byte(p.source)), Start: start, End: end}
		site.ExpressionID = compositionDigest(site)
		p.sites = append(p.sites, arithmeticProjectionSite{site, expression})
		return true
	})
}

// Recursive replacement preserves operand and short-circuit evaluation order.
// Source spans refer to the pure projection, before observation is added.
func (p arithmeticProjection) render(start, end int) string {
	var out strings.Builder
	cursor := start
	for i, site := range p.sites {
		if site.Start < cursor || site.End > end {
			continue
		}
		out.WriteString(p.source[cursor:site.Start])
		name := p.prefix + "Quotient"
		if site.Operator == "%" {
			name = p.prefix + "Remainder"
		}
		left := p.render(p.fset.Position(site.node.X.Pos()).Offset, p.fset.Position(site.node.X.End()).Offset)
		right := p.render(p.fset.Position(site.node.Y.Pos()).Offset, p.fset.Position(site.node.Y.End()).Offset)
		out.WriteString(name + "(" + left + "," + right + "," + strconv.Itoa(i) + ")")
		cursor = site.End
	}
	out.WriteString(p.source[cursor:end])
	return out.String()
}

func observedCompositionSources(prior Composition) (string, string, error) {
	projection, driver, _, err := observedCompositionArtifacts(prior)
	return projection, driver, err
}

func observedCompositionArtifacts(prior Composition) (string, string, []CompositionFaultSite, error) {
	projection, sites, err := arithmeticCompositionProjection(prior)
	if err != nil {
		return "", "", nil, err
	}
	instrumented := prior
	instrumented.Source = projection
	observer := "GoooObserveCalledInputs"
	if len(sites) != 0 {
		observer = arithmeticObservationPrefix(prior.Source) + "CalledInputs"
	}
	projection, driver, err := observedCalledCompositionWithName(instrumented, observer)
	if err == nil && len(sites) != 0 {
		driver, err = arithmeticCompositionDriver(prior)
	}
	return projection, driver, sites, err
}
