package bodyexecution

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

func compositionConditionSites(prior Composition) ([]CompositionConditionSite, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "composition.go", prior.Source, 0)
	if err != nil {
		return nil, err
	}
	bindings, sourceSHA := arithmeticFunctionIDs(prior), digest([]byte(prior.Source))
	var sites []CompositionConditionSite
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		ids := bindings[function.Name.Name]
		ast.Inspect(function.Body, func(node ast.Node) bool {
			condition, ok := node.(*ast.IfStmt)
			if !ok {
				return true
			}
			start, end := fset.Position(condition.Cond.Pos()).Offset, fset.Position(condition.Cond.End()).Offset
			site := CompositionConditionSite{RootActivityID: ids[0], ActivityID: ids[1],
				Expression: prior.Source[start:end], ProjectionSHA256: sourceSHA, Start: start, End: end}
			site.ExpressionID = compositionDigest(site)
			sites = append(sites, site)
			return true
		})
	}
	if len(sites) > 4096 {
		return nil, fmt.Errorf("native if observation exceeds 4096 condition sites")
	}
	for _, site := range sites {
		if site.ActivityID == "" || site.RootActivityID == "" {
			return nil, fmt.Errorf("native if condition has no source activity identity")
		}
	}
	return sites, nil
}

// Arithmetic observation adds calls, preserving each original if statement.
// Wrap the already lowered condition once; Go still owns short-circuit order,
// initialization scope, skipped branches, and the evaluation of helper calls.
func observedConditionProjection(source, observer string, sites []CompositionConditionSite) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "observed.go", source, parser.ParseComments)
	if err != nil {
		return "", err
	}
	index := 0
	ast.Inspect(file, func(node ast.Node) bool {
		condition, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		condition.Cond = &ast.CallExpr{Fun: ast.NewIdent(observer), Args: []ast.Expr{
			&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(index)}, condition.Cond}}
		index++
		return true
	})
	if index != len(sites) {
		return "", fmt.Errorf("native if projection differs from the retained condition inventory")
	}
	var out bytes.Buffer
	err = format.Node(&out, fset, file)
	return out.String(), err
}

func conditionObservationPrefix(source string) string {
	prefix := "GoooObservedCondition"
	for suffix := 0; strings.Contains(source, prefix); suffix++ {
		prefix = "GoooObservedCondition" + strconv.Itoa(suffix)
	}
	return prefix
}
