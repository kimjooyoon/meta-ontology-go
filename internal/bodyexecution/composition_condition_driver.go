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

// Save bounded diagnostics separately from the existing native result channel.
// Arithmetic outcomes and called inputs keep their original stdout protocol.
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
					reset := &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent(prefix + "Case"), ast.NewIdent(prefix + "Cases")}, Tok: token.ASSIGN,
						Rhs: []ast.Expr{ast.NewIdent(key.Name), &ast.CallExpr{Fun: ast.NewIdent("len"), Args: []ast.Expr{ast.NewIdent("in")}}}}
					loop.Body.List = append([]ast.Stmt{reset}, loop.Body.List...)
					ranges++
				}
			}
		}
		if call, ok := node.(*ast.CallExpr); ok && len(call.Args) == 1 {
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok && selector.Sel.Name == "Encode" {
				call.Args[0] = &ast.CallExpr{Fun: ast.NewIdent(prefix + "Save"), Args: []ast.Expr{call.Args[0]}}
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
    Conditions []GoooConditionEvent ` + "`json:\"conditions\"`" + `
    Dropped []int ` + "`json:\"dropped\"`" + `
}
type GoooConditionEvent struct {
    Case int ` + "`json:\"case_index\"`" + `
    Site int ` + "`json:\"site_index\"`" + `
    Value bool ` + "`json:\"value\"`" + `
}
var GoooConditionCase, GoooConditionCases int
var GoooConditionEvents = []GoooConditionEvent{}
var GoooConditionDropped [128]int
func GoooConditionObserve(site int, value bool) bool {
    if len(GoooConditionEvents) < 65536 {
        GoooConditionEvents = append(GoooConditionEvents,GoooConditionEvent{GoooConditionCase,site,value})
    } else { GoooConditionDropped[GoooConditionCase]++ }
    return value
}
func GoooConditionSave[T any](result T) T {
    evidence:=GoooConditionEnvelope{"gooo/native-condition-observation/v2",GoooConditionEvents,GoooConditionDropped[:GoooConditionCases]}
    raw,err:=json.Marshal(evidence)
    if err!=nil { os.Exit(4) }
    if os.WriteFile("gooo-native-condition-observation.json",raw,0600)!=nil { os.Exit(4) }
    return result
}
`
