package bodyexecution

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// Preserve the current output protocol inside an envelope. This composes with
// both called-input and arithmetic-fault observations without inventing rows.
func observedConditionDriver(source, prefix string) (string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "driver.go", source, 0)
	if err != nil {
		return "", err
	}
	ranges, encoders := 0, 0
	ast.Inspect(file, func(node ast.Node) bool {
		if loop, ok := node.(*ast.RangeStmt); ok {
			if name, ok := loop.X.(*ast.Ident); ok && name.Name == "in" {
				key, ok := loop.Key.(*ast.Ident)
				if ok {
					reset := &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(prefix + "Case")}, Tok: token.ASSIGN,
						Rhs: []ast.Expr{ast.NewIdent(key.Name)}}
					loop.Body.List = append([]ast.Stmt{reset}, loop.Body.List...)
					ranges++
				}
			}
		}
		if call, ok := node.(*ast.CallExpr); ok && len(call.Args) == 1 {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Encode" {
				call.Args[0] = &ast.CompositeLit{Type: ast.NewIdent(prefix + "Envelope"), Elts: []ast.Expr{
					&ast.BasicLit{Kind: token.STRING, Value: `"` + conditionObservationSchema + `"`},
					call.Args[0], ast.NewIdent(prefix + "Events")}}
				encoders++
			}
		}
		return true
	})
	if ranges != 1 || encoders != 1 {
		return "", fmt.Errorf("native if driver differs from the retained composition driver")
	}
	var out bytes.Buffer
	if err := format.Node(&out, fset, file); err != nil {
		return "", err
	}
	out.WriteString(strings.NewReplacer("GoooCondition", prefix).Replace(conditionObserver))
	raw, err := format.Source(out.Bytes())
	return string(raw), err
}

const conditionObserver = `
type GoooConditionEnvelope struct {
    Schema string ` + "`json:\"schema\"`" + `
    Result any ` + "`json:\"result\"`" + `
    Conditions []GoooConditionEvent ` + "`json:\"conditions\"`" + `
}
type GoooConditionEvent struct {
    Case int ` + "`json:\"case_index\"`" + `
    Site int ` + "`json:\"site_index\"`" + `
    Value bool ` + "`json:\"value\"`" + `
}
var GoooConditionCase int
var GoooConditionEvents = []GoooConditionEvent{}
func GoooConditionObserve(site int, value bool) bool {
    if len(GoooConditionEvents) >= 65536 { os.Exit(4) }
    GoooConditionEvents = append(GoooConditionEvents,GoooConditionEvent{GoooConditionCase,site,value})
    return value
}
`
