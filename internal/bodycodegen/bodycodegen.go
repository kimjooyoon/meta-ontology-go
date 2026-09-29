// Package bodycodegen lowers a deliberately small, pure activity-body language
// from .gooo source into a deterministic Go projection.
package bodycodegen

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"maps"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/decisionroute"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const (
	schema              = "gooo/body-codegen-report/v2"
	preserveRoute       = "preserve"
	guardReturnRoute    = "guard-return"
	mergeResultRoute    = "merge-result"
	routeDecisionBudget = 3 * time.Second
)

// Result is an immutable report and generated source projection.
type Result struct {
	Report Report `json:"report"`
	Source string `json:"source"`
}

// Report describes what was lowered and how the output was checked.
type Report struct {
	Schema                 string                `json:"schema"`
	Decision               string                `json:"decision"`
	Activity               string                `json:"activity"`
	ActivityID             string                `json:"activity_id"`
	SourceDigest           string                `json:"source_digest"`
	ProgramDigest          string                `json:"program_digest"`
	GeneratedDigest        string                `json:"generated_digest"`
	ReplayDigest           string                `json:"replay_digest"`
	Route                  string                `json:"route"`
	RouteDecision          decisionroute.Receipt `json:"route_decision"`
	RouteDecisionLatencyMS float64               `json:"route_decision_latency_ms"`
	CandidateRoutes        []string              `json:"candidate_routes"`
	EquivalenceRule        string                `json:"equivalence_rule"`
	SourceConstructs       int                   `json:"source_constructs"`
	LoweredConstructs      int                   `json:"lowered_constructs"`
	SourceSemanticUnits    int                   `json:"source_semantic_units"`
	LoweredSemanticUnits   int                   `json:"lowered_semantic_units"`
	CompletenessPercent    float64               `json:"completeness_percent"`
	TypecheckPassed        bool                  `json:"typecheck_passed"`
	DeterministicReplay    bool                  `json:"deterministic_replay"`
	RepositoryWrites       int                   `json:"repository_writes"`
	UnsupportedConstructs  string                `json:"unsupported_constructs"`
}

// Generate compiles one .gooo activity body into a marked Go source region.
// The accepted body subset is local declarations/assignments, conditionals,
// and returns over int64 and bool values. Calls and effects fail closed.
func Generate(filename string, source []byte, activityName string) (Result, error) {
	return GenerateWithPlanner(context.Background(), filename, source, activityName, "", "")
}

// GenerateWithPlanner asks Laya to choose only among source-shape-preserving
// lowering routes that are valid for this activity. With no usable provider,
// decisionroute selects the declared deterministic fallback.
func GenerateWithPlanner(ctx context.Context, filename string, source []byte, activityName, endpoint, apiKey string) (Result, error) {
	file, diagnostics := syntax.ParseFile(filename, string(source))
	if diagnostics.HasErrors() {
		return Result{}, fmt.Errorf("parse .gooo source: %w", diagnostics.Error())
	}
	if file == nil || file.Package == nil {
		return Result{}, fmt.Errorf(".gooo source has no package declaration")
	}
	var activity *syntax.ActivityDecl
	for _, declaration := range file.Declarations {
		if candidate, ok := declaration.(*syntax.ActivityDecl); ok && candidate.Name == activityName {
			activity = candidate
			break
		}
	}
	if activity == nil {
		return Result{}, fmt.Errorf("activity %q was not found", activityName)
	}
	if !activity.ValueProgramPresent || activity.ValueProgram == "" {
		return Result{}, fmt.Errorf("activity %q has no computes body", activityName)
	}
	inputs, outputs := activity.Inputs, activity.Output
	if len(inputs) != 1 {
		return Result{}, fmt.Errorf("activity %q requires exactly one input, got %d", activityName, len(inputs))
	}
	inputType, ok := goTypeForEntity(inputs[0].Name)
	if !ok {
		return Result{}, fmt.Errorf("activity %q input entity %q is outside the v1 profile", activityName, inputs[0].Name)
	}
	outputType, ok := goTypeForEntity(outputs)
	if !ok {
		return Result{}, fmt.Errorf("activity %q output entity %q is outside the v1 profile", activityName, outputs)
	}
	modelDocument, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return Result{}, fmt.Errorf("lower activity identity: %w", err)
	}
	model, err := bidir.Get(modelDocument)
	if err != nil {
		return Result{}, fmt.Errorf("resolve activity identity: %w", err)
	}
	activityID := ""
	for _, node := range model.Nodes {
		if node.Kind == bidir.ActivityKind && node.Name == activityName {
			activityID = string(node.ID)
			break
		}
	}
	if activityID == "" {
		return Result{}, fmt.Errorf("activity %q has no stable semantic identity", activityName)
	}

	body, err := rewriteLetDeclarations(activity.ValueProgram)
	if err != nil {
		return Result{}, err
	}
	base, err := generateRoute(file.Package.Name, activityName, activityID, inputType, outputType, body, preserveRoute)
	if err != nil {
		return Result{}, err
	}

	routes, shape, err := candidateRoutes(file.Package.Name, activityName, body)
	if err != nil {
		return Result{}, err
	}
	selected := preserveRoute
	receipt := decisionroute.Receipt{
		Schema: decisionroute.ReceiptSchema, Mode: "deterministic_fallback", Selected: preserveRoute,
		FallbackReason: "NO_ALTERNATIVE_ROUTE", Provider: "deterministic",
	}
	candidateIDs := make([]string, 0, len(routes))
	for _, option := range routes {
		candidateIDs = append(candidateIDs, option.ID)
	}
	if len(routes) > 1 {
		request := decisionroute.Request{
			Schema: decisionroute.RequestSchema,
			State:  fmt.Sprintf("activity=%s; source_sha256=%s; program_sha256=%s; input=%s; output=%s; body_shape=%s; source_semantic_units=%d", activityName, digest(source), digest([]byte(activity.ValueProgram)), inputs[0].Name, outputs, shape, base.report.SourceSemanticUnits),
			Question: decisionroute.Question{
				ID:           "body_codegen_route",
				Instructions: "Choose one listed, semantics-preserving lowering route for the described Gooo activity shape. Do not invent code or routes. Prefer the route whose generated control flow is clearest for this shape.",
				Options:      routes,
			},
			Fallback: preserveRoute,
		}
		started := time.Now()
		decisionContext, cancel := context.WithTimeout(ctx, routeDecisionBudget)
		receipt, err = decisionroute.Resolve(decisionContext, request, endpoint, apiKey)
		cancel()
		decisionLatencyMS := float64(time.Since(started)) / float64(time.Millisecond)
		if err != nil {
			return Result{}, fmt.Errorf("select code generation route: %w", err)
		}
		selected = receipt.Selected
		base.report.RouteDecisionLatencyMS = decisionLatencyMS
	}
	result := base
	if selected != preserveRoute {
		result, err = generateRoute(file.Package.Name, activityName, activityID, inputType, outputType, body, selected)
		if err != nil {
			// A planner can choose only a declared route. Keep the source-preserving
			// route authoritative if a declared lowering unexpectedly fails.
			selected = preserveRoute
			receipt.Selected = preserveRoute
			receipt.Mode = "deterministic_fallback"
			receipt.Provider = "deterministic"
			receipt.FallbackReason = "SELECTED_ROUTE_LOWERING_FAILED"
			result = base
		}
	}
	replay, err := generateRoute(file.Package.Name, activityName, activityID, inputType, outputType, body, selected)
	if err != nil {
		return Result{}, fmt.Errorf("replay selected code generation route: %w", err)
	}
	if result.report.SourceConstructs != replay.report.SourceConstructs || result.report.LoweredConstructs != replay.report.LoweredConstructs || result.report.SourceSemanticUnits != replay.report.SourceSemanticUnits || result.report.LoweredSemanticUnits != replay.report.LoweredSemanticUnits {
		return Result{}, fmt.Errorf("internal error: replay construct count changed")
	}
	result.report.SourceDigest = digest(source)
	result.report.ProgramDigest = digest([]byte(activity.ValueProgram))
	result.report.GeneratedDigest = digest(result.source)
	result.report.ReplayDigest = digest(replay.source)
	result.report.Route = selected
	result.report.RouteDecision = receipt
	result.report.RouteDecisionLatencyMS = base.report.RouteDecisionLatencyMS
	result.report.CandidateRoutes = candidateIDs
	result.report.DeterministicReplay = result.report.GeneratedDigest == result.report.ReplayDigest
	result.report.RepositoryWrites = 0
	result.report.UnsupportedConstructs = "calls, loops, imports, external effects, multiple inputs"
	return Result{Report: result.report, Source: string(result.source)}, nil
}

func goTypeForEntity(entity string) (string, bool) {
	switch entity {
	case "Integer":
		return "int64", true
	case "Boolean":
		return "bool", true
	default:
		return "", false
	}
}

type generatedRoute struct {
	report Report
	source []byte
}

func generateRoute(packageName, activityName, activityID, inputType, outputType, body, route string) (generatedRoute, error) {
	generated, sourceConstructs, loweredConstructs, sourceUnits, loweredUnits, err := render(packageName, activityName, activityID, inputType, outputType, body, route)
	if err != nil {
		return generatedRoute{}, err
	}
	completeness := 0.0
	if sourceUnits > 0 {
		covered := min(loweredUnits, sourceUnits)
		completeness = float64(covered) * 100 / float64(sourceUnits)
	}
	equivalenceRule := "source-shape-preserving-v1"
	if route == guardReturnRoute {
		equivalenceRule = "if-return-else-return-to-guard-return-v1"
	} else if route == mergeResultRoute {
		equivalenceRule = "if-return-else-return-to-explicit-result-join-v1"
	}
	return generatedRoute{source: generated, report: Report{
		Schema: schema, Decision: "PASS", Activity: activityName, ActivityID: activityID,
		Route: route, EquivalenceRule: equivalenceRule,
		SourceConstructs: sourceConstructs, LoweredConstructs: loweredConstructs,
		SourceSemanticUnits: sourceUnits, LoweredSemanticUnits: loweredUnits,
		CompletenessPercent: completeness, TypecheckPassed: true,
	}}, nil
}

func render(packageName, activityName, activityID, inputType, outputType, body, route string) ([]byte, int, int, int, int, error) {
	wrapped := fmt.Sprintf("package %s\nfunc %s(input %s) %s {\n%s\n}\n", packageName, activityName, inputType, outputType, body)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "body.goo", wrapped, parser.AllErrors)
	if err != nil {
		return nil, 0, 0, 0, 0, fmt.Errorf("parse computes body: %w", err)
	}
	function, ok := findFunction(file, activityName)
	if !ok {
		return nil, 0, 0, 0, 0, fmt.Errorf("activity body did not produce a function")
	}
	sourceConstructs, err := validateBlock(function.Body, "input", map[string]bool{"input": true}, false)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	if !blockTerminates(function.Body) {
		return nil, 0, 0, 0, 0, fmt.Errorf("activity %q body must return on every control-flow path", activityName)
	}
	sourceUnits := semanticUnitCount(function.Body)
	if route == guardReturnRoute {
		if !lowerGuardReturn(function.Body) {
			return nil, 0, 0, 0, 0, fmt.Errorf("activity %q does not match the guard-return route shape", activityName)
		}
	} else if route == mergeResultRoute {
		if !lowerMergeResult(function.Body, outputType) {
			return nil, 0, 0, 0, 0, fmt.Errorf("activity %q does not match the merge-result route shape", activityName)
		}
	} else if route != preserveRoute {
		return nil, 0, 0, 0, 0, fmt.Errorf("unknown body-codegen route %q", route)
	}
	loweredConstructs, err := validateBlock(function.Body, "input", map[string]bool{"input": true}, route == guardReturnRoute)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	if !blockTerminates(function.Body) {
		return nil, 0, 0, 0, 0, fmt.Errorf("lowered activity %q body does not return on every control-flow path", activityName)
	}
	loweredUnits := semanticUnitCount(function.Body)
	if err := typecheck(packageName, file, fset); err != nil {
		return nil, 0, 0, 0, 0, fmt.Errorf("typecheck generated activity: %w", err)
	}

	var output bytes.Buffer
	fmt.Fprintf(&output, "package %s\n\n", packageName)
	fmt.Fprintf(&output, "//gooo:generated:start id=%q kind=\"activity\"\n", activityID)
	if err := format.Node(&output, fset, function); err != nil {
		return nil, 0, 0, 0, 0, fmt.Errorf("format generated activity: %w", err)
	}
	output.WriteByte('\n')
	fmt.Fprintf(&output, "//gooo:generated:end id=%q kind=\"activity\"\n", activityID)
	formatted, err := format.Source(output.Bytes())
	if err != nil {
		return nil, 0, 0, 0, 0, fmt.Errorf("format generated source: %w", err)
	}
	return formatted, sourceConstructs, loweredConstructs, sourceUnits, loweredUnits, nil
}

func candidateRoutes(packageName, activityName, body string) ([]decisionroute.Option, string, error) {
	options := []decisionroute.Option{{
		ID:          preserveRoute,
		Description: "Preserve the explicit source control-flow shape while lowering each accepted statement and expression directly.",
	}}
	wrapped := fmt.Sprintf("package %s\nfunc %s(input int64) int64 {\n%s\n}\n", packageName, activityName, body)
	file, err := parser.ParseFile(token.NewFileSet(), "body.goo", wrapped, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return nil, "", fmt.Errorf("inspect codegen route candidates: %w", err)
	}
	function, ok := findFunction(file, activityName)
	if !ok {
		return nil, "", fmt.Errorf("activity body did not produce a function")
	}
	shape := "general-pure-body"
	if len(file.Comments) == 0 && isGuardReturnShape(function.Body) {
		shape = "single-if-else-with-one-return-per-branch"
		options = append(options, decisionroute.Option{
			ID:          guardReturnRoute,
			Description: "For a single pure if/else whose branches each return once, emit the true-branch guard followed by the false-branch return; this removes one nesting block while preserving both outcomes.",
		})
		options = append(options, decisionroute.Option{
			ID:          mergeResultRoute,
			Description: "For a single pure if/else whose branches each return once, assign both branch values to one typed result local and return it after the conditional; this makes the control-flow join explicit.",
		})
	}
	return options, shape, nil
}

func isGuardReturnShape(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) != 1 {
		return false
	}
	conditional, ok := body.List[0].(*ast.IfStmt)
	if !ok || conditional.Init != nil || conditional.Else == nil || len(conditional.Body.List) != 1 {
		return false
	}
	if _, ok := conditional.Body.List[0].(*ast.ReturnStmt); !ok {
		return false
	}
	otherwise, ok := conditional.Else.(*ast.BlockStmt)
	if !ok || len(otherwise.List) != 1 {
		return false
	}
	_, ok = otherwise.List[0].(*ast.ReturnStmt)
	return ok
}

func lowerGuardReturn(body *ast.BlockStmt) bool {
	if !isGuardReturnShape(body) {
		return false
	}
	conditional := body.List[0].(*ast.IfStmt)
	otherwise := conditional.Else.(*ast.BlockStmt)
	fallback := otherwise.List[0]
	conditional.Else = nil
	body.List = []ast.Stmt{conditional, fallback}
	return true
}

func lowerMergeResult(body *ast.BlockStmt, outputType string) bool {
	if !isGuardReturnShape(body) {
		return false
	}
	conditional := body.List[0].(*ast.IfStmt)
	otherwise := conditional.Else.(*ast.BlockStmt)
	thenReturn := conditional.Body.List[0].(*ast.ReturnStmt)
	elseReturn := otherwise.List[0].(*ast.ReturnStmt)
	resultName := "_goooResult"
	result := ast.NewIdent(resultName)
	declaration := &ast.DeclStmt{Decl: &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{&ast.ValueSpec{
			Names: []*ast.Ident{ast.NewIdent(resultName)},
			Type:  ast.NewIdent(outputType),
		}},
	}}
	merge := &ast.IfStmt{
		Cond: conditional.Cond,
		Body: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(resultName)}, Tok: token.ASSIGN, Rhs: thenReturn.Results,
		}}},
		Else: &ast.BlockStmt{List: []ast.Stmt{&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(resultName)}, Tok: token.ASSIGN, Rhs: elseReturn.Results,
		}}},
	}
	body.List = []ast.Stmt{declaration, merge, &ast.ReturnStmt{Results: []ast.Expr{result}}}
	return true
}

func semanticUnitCount(node ast.Node) int {
	count := 0
	ast.Inspect(node, func(current ast.Node) bool {
		switch value := current.(type) {
		case ast.Stmt:
			if _, container := value.(*ast.BlockStmt); !container {
				count++
			}
		case ast.Expr:
			count++
		}
		return true
	})
	return count
}

func findFunction(file *ast.File, name string) (*ast.FuncDecl, bool) {
	for _, declaration := range file.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && function.Name.Name == name {
			return function, true
		}
	}
	return nil, false
}

func validateBlock(block *ast.BlockStmt, inputName string, inherited map[string]bool, allowGuardReturn bool) (int, error) {
	if block == nil {
		return 0, fmt.Errorf("activity body has no block")
	}
	count := 0
	locals := cloneNames(inherited)
	for _, statement := range block.List {
		count++
		switch value := statement.(type) {
		case *ast.DeclStmt:
			declaration, ok := value.Decl.(*ast.GenDecl)
			if !ok || declaration.Tok != token.VAR || len(declaration.Specs) != 1 {
				return 0, fmt.Errorf("only one local let declaration is supported")
			}
			spec, ok := declaration.Specs[0].(*ast.ValueSpec)
			if !ok || len(spec.Names) != 1 {
				return 0, fmt.Errorf("let requires one local name")
			}
			name := spec.Names[0].Name
			if spec.Type != nil {
				typeName, supported := spec.Type.(*ast.Ident)
				if name != "_goooResult" || !supported || (typeName.Name != "int64" && typeName.Name != "bool") || len(spec.Values) != 0 {
					return 0, fmt.Errorf("explicit local types are reserved for compiler-generated result joins")
				}
			} else if len(spec.Values) != 1 {
				return 0, fmt.Errorf("let requires one inferred local value")
			}
			if name == inputName || locals[name] {
				return 0, fmt.Errorf("let name %q is already bound", name)
			}
			if len(spec.Values) == 1 {
				if err := validateExpression(spec.Values[0]); err != nil {
					return 0, err
				}
			}
			locals[name] = true
		case *ast.AssignStmt:
			if value.Tok != token.ASSIGN || len(value.Lhs) != 1 || len(value.Rhs) != 1 {
				return 0, fmt.Errorf("assignment requires one existing local and one value")
			}
			name, ok := value.Lhs[0].(*ast.Ident)
			if !ok || name.Name == inputName || !locals[name.Name] {
				return 0, fmt.Errorf("assignment target must be an existing local")
			}
			if err := validateExpression(value.Rhs[0]); err != nil {
				return 0, err
			}
		case *ast.IfStmt:
			if value.Init != nil || (value.Else == nil && !allowGuardReturn) {
				return 0, fmt.Errorf("if requires a condition and an explicit else branch")
			}
			if err := validateExpression(value.Cond); err != nil {
				return 0, err
			}
			thenCount, err := validateBlock(value.Body, inputName, locals, allowGuardReturn)
			if err != nil {
				return 0, err
			}
			count += thenCount
			if value.Else == nil {
				continue
			}
			switch otherwise := value.Else.(type) {
			case *ast.BlockStmt:
				elseCount, err := validateBlock(otherwise, inputName, locals, allowGuardReturn)
				if err != nil {
					return 0, err
				}
				count += elseCount
			case *ast.IfStmt:
				elseCount, err := validateBlock(&ast.BlockStmt{List: []ast.Stmt{otherwise}}, inputName, locals, allowGuardReturn)
				if err != nil {
					return 0, err
				}
				count += elseCount
			default:
				return 0, fmt.Errorf("else branch must be a block or if")
			}
		case *ast.ReturnStmt:
			if len(value.Results) != 1 {
				return 0, fmt.Errorf("return requires exactly one value")
			}
			if err := validateExpression(value.Results[0]); err != nil {
				return 0, err
			}
		default:
			return 0, fmt.Errorf("unsupported activity statement %T", statement)
		}
	}
	return count, nil
}

func cloneNames(names map[string]bool) map[string]bool {
	clone := make(map[string]bool, len(names))
	maps.Copy(clone, names)
	return clone
}

func validateExpression(expression ast.Expr) error {
	switch value := expression.(type) {
	case *ast.Ident, *ast.BasicLit:
		return nil
	case *ast.ParenExpr:
		return validateExpression(value.X)
	case *ast.UnaryExpr:
		if value.Op != token.SUB && value.Op != token.NOT {
			return fmt.Errorf("unsupported unary operator %s", value.Op)
		}
		return validateExpression(value.X)
	case *ast.BinaryExpr:
		switch value.Op {
		case token.ADD, token.SUB, token.MUL, token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ, token.LAND, token.LOR:
		default:
			return fmt.Errorf("unsupported binary operator %s", value.Op)
		}
		if err := validateExpression(value.X); err != nil {
			return err
		}
		return validateExpression(value.Y)
	default:
		return fmt.Errorf("unsupported activity expression %T", expression)
	}
}

func blockTerminates(block *ast.BlockStmt) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.IfStmt:
		if last.Else == nil || !blockTerminates(last.Body) {
			return false
		}
		switch otherwise := last.Else.(type) {
		case *ast.BlockStmt:
			return blockTerminates(otherwise)
		case *ast.IfStmt:
			return blockTerminates(&ast.BlockStmt{List: []ast.Stmt{otherwise}})
		}
	}
	return false
}

func typecheck(packageName string, file *ast.File, fset *token.FileSet) error {
	configuration := types.Config{Importer: importer.Default()}
	_, err := configuration.Check(packageName, fset, []*ast.File{file}, nil)
	return err
}

func rewriteLetDeclarations(body string) (string, error) {
	fset := token.NewFileSet()
	file := fset.AddFile("computes", fset.Base(), len(body))
	var sourceScanner scanner.Scanner
	sourceScanner.Init(file, []byte(body), nil, scanner.ScanComments)
	var offsets []int
	for {
		position, kind, literal := sourceScanner.Scan()
		if kind == token.EOF {
			break
		}
		if kind == token.IDENT && literal == "let" {
			offsets = append(offsets, file.Offset(position))
		}
	}
	if len(offsets) == 0 {
		return body, nil
	}
	var output strings.Builder
	last := 0
	for _, offset := range offsets {
		if offset < last || offset+3 > len(body) {
			return "", fmt.Errorf("invalid let token in computes body")
		}
		output.WriteString(body[last:offset])
		output.WriteString("var")
		last = offset + 3
	}
	output.WriteString(body[last:])
	return output.String(), nil
}

func digest(value []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(value)) }
