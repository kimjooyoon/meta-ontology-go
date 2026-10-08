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

// Native observation is derived after source replay. The saved pure projection
// remains unchanged; the runtime binds the separately instrumented source hashes.
func observedCalledCompositionSources(prior Composition) (string, string, error) {
	return observedCalledCompositionWithName(prior, "GoooObserveCalledInputs")
}

func observedCalledCompositionWithName(prior Composition, observerName string) (string, string, error) {
	if len(prior.Preparations) == 0 {
		return prior.Source, prior.Driver, nil
	}
	helpers := make(map[string]bool, len(prior.Preparations))
	for _, step := range prior.Preparations {
		helpers[step.Generation.Report.ActivityID] = true
	}
	type binding struct{ root, helper string }
	names := map[string]binding{}
	for i, step := range prior.Steps {
		if closure := step.Generation.Report.CallClosure; closure != nil {
			for c, helper := range closure.Activities {
				if helpers[helper.ActivityID] {
					names[fmt.Sprintf("GoooComposedActivity%dCall%d", i, c)] = binding{step.Generation.Report.ActivityID, helper.ActivityID}
				}
			}
		}
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "observed.go", prior.Source, 0)
	if err != nil {
		return "", "", err
	}
	count := 0
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		binding, observe := names[function.Name.Name]
		if !observe {
			continue
		}
		arguments := []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(binding.root)},
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(binding.helper)}}
		for _, field := range function.Type.Params.List {
			for _, name := range field.Names {
				arguments = append(arguments, ast.NewIdent(name.Name))
			}
		}
		statement := &ast.ExprStmt{X: &ast.CallExpr{Fun: ast.NewIdent(observerName), Args: arguments}}
		function.Body.List = append([]ast.Stmt{statement}, function.Body.List...)
		count++
	}
	if count != len(names) {
		return "", "", fmt.Errorf("observed call projection differs from the replayed closure")
	}
	var projection bytes.Buffer
	if err := format.Node(&projection, fset, file); err != nil {
		return "", "", err
	}
	driver, err := observedCompositionDriver(prior.Driver)
	return projection.String(), driver, err
}

func observedCompositionDriver(source string) (string, error) {
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
				if !ok {
					return false
				}
				assignment := &ast.AssignStmt{Lhs: []ast.Expr{ast.NewIdent("goooObservationCase"), ast.NewIdent("goooObservationCount")},
					Tok: token.ASSIGN, Rhs: []ast.Expr{ast.NewIdent(key.Name), &ast.BasicLit{Kind: token.INT, Value: "0"}}}
				loop.Body.List = append([]ast.Stmt{assignment}, loop.Body.List...)
				ranges++
			}
		}
		if call, ok := node.(*ast.CallExpr); ok && len(call.Args) == 1 {
			selector, ok := call.Fun.(*ast.SelectorExpr)
			name, output := call.Args[0].(*ast.Ident)
			if ok && output && selector.Sel.Name == "Encode" && name.Name == "out" {
				expression, parseErr := parser.ParseExpr(`GoooObservedComposition{Schema:"gooo/called-input-observation/v1",Outputs:out,Calls:goooObservationCalls}`)
				if parseErr != nil {
					panic(parseErr)
				}
				call.Args[0] = expression
				encoders++
			}
		}
		return true
	})
	if ranges != 1 || encoders != 1 {
		return "", fmt.Errorf("observed input driver differs from the replayed driver")
	}
	var driver strings.Builder
	if err := format.Node(&driver, fset, file); err != nil {
		return "", err
	}
	driver.WriteString(calledInputObserver)
	raw, err := format.Source([]byte(driver.String()))
	return string(raw), err
}

const calledInputObserver = `
type GoooObservedComposition struct {
    Schema string ` + "`json:\"schema\"`" + `
    Outputs any ` + "`json:\"outputs\"`" + `
    Calls []GoooObservedCall ` + "`json:\"calls\"`" + `
}
type GoooObservedCall struct {
    Case int ` + "`json:\"case_index\"`" + `
    Root string ` + "`json:\"root_activity_id\"`" + `
    Activity string ` + "`json:\"activity_id\"`" + `
    Inputs []json.RawMessage ` + "`json:\"inputs\"`" + `
}
var goooObservationCase, goooObservationCount int
var goooObservationCalls = []GoooObservedCall{}
func GoooObserveCalledInputs(root, activity string, inputs ...any) {
    if goooObservationCount >= 65536 { os.Exit(4) }
    goooObservationCount++
    values := make([]json.RawMessage,len(inputs))
    for i, input := range inputs {
        value, err := json.Marshal(input)
        if err != nil { os.Exit(4) }
        values[i] = value
    }
    goooObservationCalls = append(goooObservationCalls,GoooObservedCall{goooObservationCase,root,activity,values})
}
`
