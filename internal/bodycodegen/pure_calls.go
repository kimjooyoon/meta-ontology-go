package bodycodegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const pureCallLimit = 4096

func bodyPrimitiveName(name string) bool { return name == "len" || name == "int64" }

// IsBodyPrimitiveName identifies the pure body's primitive spellings. A source
// activity or local value with the same name still takes precedence in scope.
func IsBodyPrimitiveName(name string) bool { return bodyPrimitiveName(name) }

type PureCallActivity struct {
	Name          string `json:"name"`
	ActivityID    string `json:"activity_id"`
	ProgramSHA256 string `json:"program_sha256"`
}

type PureCallClosure struct {
	Schema                string             `json:"schema"`
	Activities            []PureCallActivity `json:"activities"`
	Edges                 []PureCallEdge     `json:"edges"`
	MaxCallsPerInvocation int                `json:"max_calls_per_invocation"`
}

type PureCallEdge struct {
	CallerID string `json:"caller_id"`
	CalleeID string `json:"callee_id"`
	Ordinal  int    `json:"ordinal"`
}

type pureCallFunction struct {
	identity   PureCallActivity
	parameters []InputParameter
	output     string
	body       string
}

func (f pureCallFunction) declaration() string {
	return fmt.Sprintf("func %s(%s) %s {\n%s\n}\n", f.identity.Name, parameterDeclaration(f.parameters), f.output, f.body)
}

type pureCallResolver struct {
	file      *syntax.File
	records   []RecordType
	ids       map[string]string
	visiting  map[string]bool
	budgets   map[string]int
	heights   map[string]int
	functions []pureCallFunction
	edges     []PureCallEdge
}

func directPureCalls(body string) ([]string, error) {
	file, err := parser.ParseFile(token.NewFileSet(), "calls.go", "package p\nfunc body(){\n"+body+"\n}", parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var names []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || err != nil {
			return err == nil
		}
		if err = validatePureCallExpression(call); err != nil {
			return false
		}
		names = append(names, call.Fun.(*ast.Ident).Name)
		return true
	})
	return names, err
}

func validatePureCallExpression(call *ast.CallExpr) error {
	if _, ok := call.Fun.(*ast.Ident); !ok || call.Ellipsis.IsValid() || len(call.Args) < 1 || len(call.Args) > 16 {
		return fmt.Errorf("pure activity call requires a declared name and 1..16 positional arguments")
	}
	for _, argument := range call.Args {
		if err := validateExpression(argument); err != nil {
			return err
		}
	}
	return nil
}

func (r *pureCallResolver) visit(name, body string, depth int) (int, error) {
	if r.visiting[name] {
		return 0, fmt.Errorf("recursive pure activity call at %q", name)
	}
	if budget, ok := r.budgets[name]; ok {
		if depth+r.heights[name] > 16 {
			return 0, fmt.Errorf("pure activity closure exceeds 16 call levels")
		}
		return budget, nil
	}
	if depth > 16 || len(r.budgets)+len(r.visiting) >= 32 {
		return 0, fmt.Errorf("pure activity closure exceeds 32 activities or 16 call levels")
	}
	r.visiting[name] = true
	defer delete(r.visiting, name)
	names, err := directPureCalls(body)
	if err != nil {
		return 0, err
	}
	budget, height := 0, 0
	for ordinal, callee := range names {
		// Source-declared activities retain their identity even when their name
		// shadows a Go primitive. Primitives themselves add no activity edge.
		if bodyPrimitiveName(callee) && r.ids[callee] == "" {
			continue
		}
		cost, err := r.include(callee, depth+1)
		if err != nil {
			return 0, err
		}
		r.edges = append(r.edges, PureCallEdge{CallerID: r.ids[name], CalleeID: r.ids[callee], Ordinal: ordinal})
		height = max(height, 1+r.heights[callee])
		budget += 1 + cost
		if budget > pureCallLimit {
			return 0, fmt.Errorf("pure activity closure exceeds %d calls per invocation", pureCallLimit)
		}
	}
	r.budgets[name] = budget
	r.heights[name] = height
	return budget, nil
}

func (r *pureCallResolver) include(name string, depth int) (int, error) {
	if r.visiting[name] {
		return 0, fmt.Errorf("recursive pure activity call at %q", name)
	}
	if _, ok := r.budgets[name]; ok {
		return r.visit(name, "", depth)
	}
	activity, err := sourceBodyActivity(r.file, name)
	if err != nil || activity.Assembly != nil || r.ids[name] == "" {
		return 0, fmt.Errorf("pure call %q requires a fixed computes activity in the same source", name)
	}
	parameters, err := sourceBodyParameters(activity, r.records)
	if err != nil {
		return 0, err
	}
	output, ok := bodyEntityType(activity.Output, r.records)
	if !ok {
		return 0, fmt.Errorf("pure call %q output is outside the value profile", name)
	}
	body, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return 0, err
	}
	budget, err := r.visit(name, body, depth)
	if err == nil {
		r.functions = append(r.functions, pureCallFunction{identity: PureCallActivity{
			Name: name, ActivityID: r.ids[name], ProgramSHA256: digest([]byte(activity.ValueProgram))},
			parameters: parameters, output: output, body: body})
	}
	return budget, err
}

func (p preparedBody) generateBody(body, route string) (generatedRoute, error) {
	names, err := directPureCalls(body)
	if err != nil {
		return generatedRoute{}, err
	}
	if len(names) == 0 {
		return generateRouteParameters(p.packageName, p.activity.Name, p.activityID, p.parameters, p.outputType, body, route, p.records...)
	}
	functions, budget, edges, err := p.resolvePureCalls(body)
	if err != nil {
		return generatedRoute{}, err
	}
	result, err := generateRouteWithCalls(p.packageName, p.activity.Name, p.activityID,
		p.parameters, p.outputType, body, route, functions, p.allRecords)
	if err == nil && len(functions) > 0 {
		result.report.CallClosure = &PureCallClosure{Schema: "gooo/pure-activity-call-closure/v1", MaxCallsPerInvocation: budget, Edges: edges}
		for _, f := range functions {
			result.report.CallClosure.Activities = append(result.report.CallClosure.Activities, f.identity)
		}
	}
	return result, err
}

func (p preparedBody) resolvePureCalls(body string) ([]pureCallFunction, int, []PureCallEdge, error) {
	r := pureCallResolver{file: p.file, records: p.allRecords, ids: p.activityIDs,
		visiting: map[string]bool{}, budgets: map[string]int{}, heights: map[string]int{}}
	budget, err := r.visit(p.activity.Name, body, 0)
	return r.functions, budget, r.edges, err
}
